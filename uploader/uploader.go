package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	cli "github.com/jawher/mow.cli"
	util "github.com/kopoli/go-util"
)

type Config struct {
	opts     util.Options
	files    []string
	jobCount int
	timeout  int
}

func Cli(c *Config, args []string) (err error) {
	progName := c.opts.Get("program-name", "paperless-uploader")
	progVersion := c.opts.Get("program-version", "undefined")
	app := cli.App(progName, "Paperless Uploader")

	app.Version("version v", fmt.Sprintf("%s: %s\nBuilt with: %s/%s on %s/%s",
		progName, progVersion, runtime.Compiler, runtime.Version(),
		runtime.GOOS, runtime.GOARCH))

	app.Spec = "[OPTIONS] URL FILES..."

	optTags := app.StringOpt("t tag", "", "Comma separated list of tags.")
	optJobs := app.IntOpt("j jobs", runtime.NumCPU(), "Number of concurrent uploads")
	optTimeout := app.IntOpt("timeout", 60, "HTTP/S timeout in seconds")
	argURL := app.StringArg("URL", "", "The upload URL.")
	argFiles := app.StringsArg("FILES", []string{}, "Image files to upload.")

	app.Action = func() {
		c.opts.Set("tags", *optTags)
		c.opts.Set("url", *argURL)
		c.jobCount = *optJobs
		c.timeout = *optTimeout

		c.files = *argFiles
	}

	err = app.Run(args)
	if err != nil {
		return
	}

	return
}

func CheckArguments(c *Config) (err error) {
	var u *url.URL

	urlstr := c.opts.Get("url", "")
	u, err = url.Parse(urlstr)
	if err != nil {
		return
	}

	if !u.IsAbs() {
		err = util.E.New("Supplied URL must be absolute: %s", urlstr)
		return
	}

	for i := range c.files {
		var st os.FileInfo
		st, err = os.Stat(c.files[i])
		if err != nil || !st.Mode().IsRegular() {
			err = util.E.New("Invalid file: %s", c.files[i])
			return
		}
	}

	return
}

func UploadFile(c *Config, file string) (err error) {
	fp, err := os.Open(file)
	if err != nil {
		return
	}
	defer fp.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filepath.Base(file))
	if err != nil {
		return
	}

	_, err = io.Copy(part, fp)
	if err != nil {
		return
	}

	err = writer.WriteField("tags", c.opts.Get("tags", ""))
	if err != nil {
		return
	}
	err = writer.Close()
	if err != nil {
		return
	}

	url := c.opts.Get("url", "") + "/api/v1/image"

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: time.Duration(c.timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body.Reset()
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return
	}
	resp.Body.Close()

	fmt.Println("Uploaded:", file, "Response:", body.String())

	if resp.StatusCode != http.StatusCreated {
		err = util.E.New("Server responded unexpectedly with code: %d", resp.StatusCode)
		return
	}

	return
}

func main() {
	ret := 0
	config := &Config{
		opts: util.NewOptions(),
	}

	config.opts.Set("program-name", os.Args[0])

	err := Cli(config, os.Args)
	if err != nil {
		err = util.E.Annotate(err, "Command line parsing failed")
		goto error
	}

	err = CheckArguments(config)
	if err != nil {
		err = util.E.Annotate(err, "Invalid arguments")
		goto error
	}

	fmt.Println("Files:", config.files)

	for i := range config.files {
		err = UploadFile(config, config.files[i])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Uploading file:", config.files[i],"failed with:",err)
			ret = 1
		}
	}

	os.Exit(ret)

error:
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
