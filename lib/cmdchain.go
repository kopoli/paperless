package paperless

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"github.com/kopoli/go-util"
)

// Environment is the run environment for each command. It is supplied as part
// of Status for a Link when it is Run or Validated.
type Environment struct {

	// Constants are variables that are defined before a command chain is run
	Constants map[string]string

	// Tempfiles are constants that house a name of a temporary file. The
	// files are created before the chain is run and they are removed at
	// the end.
	TempFiles []string

	// The directory where the commands are run. This is a created
	// temporary directory
	RootDir string

	// AllowedCommands contain the commands that are allowed. If this is
	// nil, all commands are allowed.
	AllowedCommands map[string]bool

	initialized bool
}

func (e *Environment) initEnv() (err error) {
	if e.initialized {
		return
	}

	if e.Constants == nil {
		return util.E.New("field Constants not initialized")
	}

	e.RootDir, err = ioutil.TempDir("", "chain")
	if err != nil {
		return util.E.Annotate(err, "rootdir creation failed")
	}

	var fp *os.File
	for _, name := range e.TempFiles {
		fp, err = ioutil.TempFile(e.RootDir, "tmp")
		if err != nil {
			err = util.E.Annotate(err, "tempfile creation failed")
			e.initialized = true
			e2 := e.deinitEnv()
			if e2 != nil {
				err = util.E.Annotate(err, "temproot removal failed:", e2)
				e.initialized = false
			}
			return
		}
		e.Constants[name] = fp.Name()
		fp.Close()
	}

	e.initialized = true
	return
}

func (e *Environment) deinitEnv() (err error) {
	if !e.initialized {
		return
	}

	if !strings.HasPrefix(e.RootDir, os.TempDir()) || e.RootDir == os.TempDir() {
		err = util.E.New("Temporary directory path is corrupted: %s", e.RootDir)
		return
	}

	err = os.RemoveAll(e.RootDir)
	if err != nil {
		err = util.E.Annotate(err, "tempdir removal failed:")
	}

	// Clear the temporary file and directory names
	e.RootDir = ""
	for _, n := range e.TempFiles {
		e.Constants[n] = ""
	}

	e.initialized = false
	return
}

func (e *Environment) validate() (err error) {
	if len(e.RootDir) == 0 {
		return util.E.New("the RootDir must be defined")
	}
	info, err := os.Stat(e.RootDir)
	if err != nil || info.Mode()&os.ModeDir == 0 {
		return util.E.Annotate(err, "file ", e.RootDir, " is not a proper directory")
	}

	return
}

// Status is the runtime status of the command chain
type Status struct {
	// The log output will be written to this
	Log io.Writer

	Environment
}

type Link interface {
	Validate(*Environment) error
	Run(*Status) error
}

type CmdChain struct {

	//TODO remove this (this should come from outside)
	Environment

	Links []Link
}

func (c *CmdChain) Validate(e *Environment) (err error) {
	for _, l := range c.Links {
		err = l.Validate(e)
		if err != nil {
			return
		}
	}
	return
}

func (c *CmdChain) Run(s *Status) (err error) {
	err = c.Validate(&s.Environment)
	if err != nil {
		return
	}

	for i := range c.Links {
		err = c.Links[i].Run(s)
		if err != nil {
			return
		}
	}

	return
}

func RunCmdChain(c *CmdChain, s *Status) (err error) {
	err = s.Environment.initEnv()
	if err != nil {
		return
	}

	err = c.Run(s)
	e2 := s.Environment.deinitEnv()
	if e2 != nil {
		err = util.E.Annotate(err, "cmdchain deinit failed: ", e2)
	}

	return
}

////////////////////////////////////////////////////////////

type Cmd struct {
	Cmd []string
}

// NewCmd creates a new Cmd from given command string
func NewCmd(cmdstr string) (c *Cmd, err error) {
	command := splitWsQuote(cmdstr)

	if len(command) == 0 {
		return nil, util.E.New("A command could not be parsed from: %s", cmdstr)
	}

	c = &Cmd{command}

	_, err = exec.LookPath(c.Cmd[0])
	if err != nil {
		return nil, util.E.Annotate(err, "Command", c.Cmd[0], "could not be found")

	}

	return
}

// Validate makes sure the command is proper and can be run
func (c *Cmd) Validate(e *Environment) (err error) {
	if len(c.Cmd) == 0 {
		return util.E.New("command string must be non-empty")
	}

	if e.AllowedCommands != nil {
		if _, ok := e.AllowedCommands[c.Cmd[0]]; ok != true {
			return util.E.New("command is not allowed")
		}
	}

	_, err = exec.LookPath(c.Cmd[0])
	if err != nil {
		return
	}

	err = e.validate()

	for idx, a := range c.Cmd {
		consts := parseConsts(a)
		if len(consts) > 0 {
			for _, co := range consts {
				if _, ok := e.Constants[co]; !ok {
					return util.E.New("constant \"%s\" not defined", co)
				}
			}
		}

		// Output redirection to a file
		if a == ">" && (idx == len(c.Cmd)-1 || c.Cmd[idx+1] == "") {
			return util.E.New("The output redirection requires a string")
		}
	}

	return
}

func (c *Cmd) Run(s *Status) (err error) {
	err = c.Validate(&s.Environment)
	if err != nil {
		return
	}

	var args []string
	for i := range c.Cmd {
		args = append(args, expandConsts(c.Cmd[i], s.Constants))
	}

	if s.Log != nil {
		fmt.Fprintln(s.Log, "# Running command:", strings.Join(args, " "))
	}

	var output io.Writer = s.Log

	redirout, pos := getRedirectFile(">", args)
	if redirout != "" {
		var fp *os.File
		redirout = PathAbs(s.RootDir, redirout)
		fp, err = os.OpenFile(redirout, os.O_WRONLY | os.O_CREATE, 0666)
		if err != nil {
			err = util.E.Annotate(err, "Could not open file",redirout,"for redirection")
			return
		}
		defer fp.Close()
		output = fp

		// Remove the redirection and the file argument from the command
		args = append(args[:pos], args[pos+2:]...)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = s.RootDir
	cmd.Stdout = output
	cmd.Stderr = s.Log

	return cmd.Run()
}

////////////////////////////////////////////////////////////

var (
	constRe         = regexp.MustCompile(`\$(\w+)`)
	tmpfileConstRe  = regexp.MustCompile(`\$(tmp\w+)`)
	commentRe       = regexp.MustCompile(`#.*$`)
	preWhitespaceRe = regexp.MustCompile(`^\s+`)
)

// parseConsts parses the constants from a string. Returns a list of constant names
func parseConsts(s string) (ret []string) {
	ret = []string{}

	matches := constRe.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return
	}
	for _, m := range matches {
		ret = append(ret, m[1])
	}

	return
}

func expandConsts(s string, constants map[string]string) string {
	return constRe.ReplaceAllStringFunc(s, func(match string) string {
		cs := parseConsts(match)
		if len(cs) != 1 {
			panic("Invalid Regexp parsing")
		}

		ret, ok := constants[cs[0]]
		if !ok {
			ret = ""
		}

		return ret
	})
}

// Gets the string after the given redir string. If not found, returns empty
// string.
func getRedirectFile(redir string, args []string) (file string, pos int) {
	for i := range args {
		if args[i] == redir {
			if i+1 < len(args) {
				file = args[i+1]
				pos = i
			}
			return
		}
	}
	return
}

// NewCmdChainScript creates a CmdChain from a script where each command is on a separate line. The following syntax elements are supported:
//
// - Empty lines are filtered out.
//
// - Comments start with # and end with EOL.
//
// - Constants are strings that begin with $ and they can be set before running the cmdchain.
//
// - Temporary files are strings that start with $tmp and they are automatically created before running the cmdchain and removed afterwards.
func NewCmdChainScript(script string) (c *CmdChain, err error) {
	c = &CmdChain{}
	c.Constants = make(map[string]string)

	for _, line := range strings.Split(script, "\n") {
		line = commentRe.ReplaceAllString(line, "")
		line = preWhitespaceRe.ReplaceAllString(line, "")

		if len(line) == 0 {
			continue
		}

		constants := parseConsts(line)
		for _, co := range constants {
			c.Constants[co] = ""
			if tmpfileConstRe.MatchString("$" + co) {
				c.TempFiles = append(c.TempFiles, co)
			}
		}

		var cmd *Cmd
		cmd, err = NewCmd(line)
		if err != nil {
			return nil, util.E.Annotate(err, "improper command")
		}

		c.Links = append(c.Links, cmd)
	}

	e := c.Environment
	e.RootDir = "/"

	err = c.Validate(&e)
	if err != nil {
		return nil, util.E.Annotate(err, "invalid command chain")
	}

	return
}

// splitWsQuote splits a string by whitespace, but takes doublequotes into
// account
func splitWsQuote(s string) []string {

	quote := rune(0)

	return strings.FieldsFunc(s, func(r rune) bool {
		switch {
		case r == quote:
			quote = rune(0)
			return true
		case quote != rune(0):
			return false
		case unicode.In(r, unicode.Quotation_Mark):
			quote = r
			return true
		default:
			return unicode.IsSpace(r)
		}
	})
}
