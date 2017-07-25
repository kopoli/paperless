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
	"sync"
	"time"

	cli "github.com/jawher/mow.cli"
	util "github.com/kopoli/go-util"
)

const (
	APIPath string = "/api/v1/image"
)

type Config struct {
	opts     util.Options
	files    []string
	jobCount int
	timeout  int
}

func runCli(c *Config, args []string) (err error) {
	progName := c.opts.Get("program-name", "paperless-uploader")
	progVersion := c.opts.Get("program-version", "undefined")
	app := cli.App(progName, "Upload tool to Paperless Office server.")

	app.Version("version", fmt.Sprintf("%s: %s\nBuilt with: %s/%s on %s/%s",
		progName, progVersion, runtime.Compiler, runtime.Version(),
		runtime.GOOS, runtime.GOARCH))

	app.Spec = "[OPTIONS] URL FILES..."

	optJobs := app.IntOpt("j jobs", runtime.NumCPU(), "Number of concurrent uploads")
	optVerbose := app.BoolOpt("v verbose", false, "Print upload statuses")
	optTags := app.StringOpt("t tag", "", "Comma separated list of tags.")
	optTimeout := app.IntOpt("timeout", 60, "HTTP timeout in seconds")
	argURL := app.StringArg("URL", "", "The upload HTTP URL.")
	argFiles := app.StringsArg("FILES", []string{}, "Image files to upload.")

	app.Action = func() {
		c.opts.Set("tags", *optTags)
		if *optVerbose {
			c.opts.Set("verbose", "t")
		}
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

func checkArguments(c *Config) (err error) {
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

func uploadFile(c *Config, file string) (err error) {
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

	apipath, err := url.Parse(APIPath)
	if err != nil {
		return
	}
	base, err := url.Parse(c.opts.Get("url", ""))
	if err != nil {
		return
	}

	urlstr := base.ResolveReference(apipath).String()
	req, err := http.NewRequest("POST", urlstr, body)
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

	if c.opts.IsSet("verbose") {
		fmt.Println("Uploaded to url:", urlstr)
		fmt.Println("Uploaded:", file, "Response:", body.String())
	}

	if resp.StatusCode != http.StatusCreated {
		err = util.E.New("Server responded unexpectedly with code: %d", resp.StatusCode)
		return
	}

	return
}

func upload(c *Config) (err error) {
	jobs := make(chan string, 10)
	wg := sync.WaitGroup{}
	worker := func(jobs <-chan string) {
		var err error
		for file := range jobs {
			err = uploadFile(c, file)
			if err != nil {
				fmt.Printf("%s: failed: %s\n", file, err)
			} else {
				fmt.Printf("%s: Uploaded ok\n", file)
			}
		}
		wg.Done()
	}

	for i := 0; i < c.jobCount; i++ {
		wg.Add(1)
		go worker(jobs)
	}

	for i := range c.files {
		jobs <- c.files[i]
	}

	close(jobs)
	wg.Wait()

	return nil
}

func main() {
	config := &Config{
		opts: util.NewOptions(),
	}

	config.opts.Set("program-name", os.Args[0])

	err := runCli(config, os.Args)
	if err != nil {
		err = util.E.Annotate(err, "Command line parsing failed")
		goto error
	}

	err = checkArguments(config)
	if err != nil {
		err = util.E.Annotate(err, "Invalid arguments")
		goto error
	}

	err = upload(config)
	if err != nil {
		goto error
	}

	os.Exit(0)

error:
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
