package paperless

import (
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
	RootDir   string

	// AllowedCommands contain the commands that are allowed. If this is
	// nil, all commands are allowed.
	AllowedCommands map[string]bool
}

type Status struct {
	LastErr error

	Environment
}

type Link interface {
	Validate(Status) error
	Run(Status) error
}

type CmdChain struct {
	Environment
	Links []Link
}

func (c *CmdChain) Validate(Status) error {
	panic("not implemented")
}

func (c *CmdChain) Run(Status) error {
	panic("not implemented")
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
func (c *Cmd) Validate(s Status) (err error) {
	if s.LastErr != nil {
		return s.LastErr
	}

	if len(c.Cmd) == 0 {
		return util.E.New("command string must be non-empty")
	}

	if s.AllowedCommands != nil {
		if _, ok := s.AllowedCommands[c.Cmd[0]]; ok != true {
			return util.E.New("command is not allowed")
		}
	}

	_, err = exec.LookPath(c.Cmd[0])
	if err != nil {
		return
	}

	for _, a := range c.Cmd {
		consts := parseConsts(a)
		if len(consts) > 0 {
			for _, co := range consts {
				if _, ok := s.Constants[co]; !ok {
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

func NewCmdChainScript(script string) (c *CmdChain, err error) {
	c = &CmdChain{}

	for _, line := range strings.Split(script, "\n") {
		line = commentRe.ReplaceAllString(line, "")
		line = preWhitespaceRe.ReplaceAllString(line, "")

		if len(line) == 0 {
			continue
		}

		// Initialize the added constants
		constants := parseConsts(line)
		if len(constants) > 0 {
			if c.Constants == nil {
				c.Constants = make(map[string]string)
			}

			for _, co := range constants {
				c.Constants[co] = ""
			}
		}

		cmd, err := NewCmd(line)
		if err != nil {
			return nil, util.E.Annotate(err, "Improper command")
		}

		err = cmd.Validate(Status{Environment: c.Environment})
		if err != nil {
			return nil, err
		}

		c.Links = append(c.Links, cmd)
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
