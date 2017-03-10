package paperless

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"github.com/kopoli/go-util"
)

// TODO:
// - Validation:
//   - The constants are defined.
//   - The programs are found.

// Environment is the run environment for each command. It is supplied as part
// of Status for a Link when it is Run or Validated.
type Environment struct {

	// Constants are variables that are defined before a command chain is run
	Constants map[string]string

	// Tempfiles are constants that house a name of a temporary file. The
	// files are created before the chain is run and they are removed at
	// the end.
	TempFiles []string

	// The directory where the commands are run.
	RootDir string

	// AllowedCommands contain the commands that are allowed. If this is
	// nil, all commands are allowed.
	AllowedCommands map[string]bool
}

type Status struct {
	LastErr error

	Environment
}

type Link interface {
	Validate(Environment) error
	Run(Status) error
}

type CmdChain struct {
	Environment
	Links []Link
}

func (c *CmdChain) Validate(e Environment) (err error) {
	for _, l := range c.Links {
		err = l.Validate(e)
		if err != nil {
			return
		}
	}
	return
}

func (c *CmdChain) Run(s Status) (err error) {
	err = c.Validate(s.Environment)
	if err != nil {
		return
	}

	// TODO
	// - make sure the constants are defined
	// - Create the temporary files
	// - Create the run directory

	panic("not implemented")
}

func RunCmdChain(c *CmdChain, constants map[string]string) (err error) {
	s := Status{Environment: c.Environment}
	s.Constants = constants

	return c.Run(s)
}

type Cmd struct {
	Cmd []string
}

// NewCmd creates a new Cmd from given command string
func NewCmd(cmdstr string) (c *Cmd, err error) {
	command := splitWsQuote(cmdstr)

	if len(command) == 0 {
		return nil, util.E.New("A command could not be parsed from:", cmdstr)
	}

	c = &Cmd{command}

	_, err = exec.LookPath(c.Cmd[0])
	if err != nil {
		return nil, util.E.Annotate(err, "Command", c.Cmd[0], "could not be found")

	}

	return
}

// Validate makes sure the command is proper and can be run
func (c *Cmd) Validate(e Environment) (err error) {
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

	if len(e.RootDir) == 0 {
		return util.E.New("the RootDir must be defined")
	}
	info, err := os.Stat(e.RootDir)
	if err != nil || info.Mode()&os.ModeDir == 0 {
		return util.E.Annotate(err, "file ", e.RootDir, " is not a proper directory")
	}

	for _, a := range c.Cmd {
		consts := parseConsts(a)
		if len(consts) > 0 {
			for _, co := range consts {
				if _, ok := e.Constants[co]; !ok {
					return util.E.New("constant \"%s\" not defined", co)
				}
			}
		}
	}

	return
}

func (c *Cmd) Run(Status) error {
	panic("not implemented")
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

	for _, line := range strings.Split(script, "\n") {
		line = commentRe.ReplaceAllString(line, "")
		line = preWhitespaceRe.ReplaceAllString(line, "")

		if len(line) == 0 {
			continue
		}

		constants := parseConsts(line)
		if len(constants) > 0 {
			if c.Constants == nil {
				c.Constants = make(map[string]string)
			}

			for _, co := range constants {
				c.Constants[co] = ""
				if tmpfileConstRe.MatchString("$" + co) {
					c.TempFiles = append(c.TempFiles, co)
				}
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

	err = c.Validate(e)
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
