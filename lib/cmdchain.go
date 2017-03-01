package paperless

import (
	"errors"
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

type Environment struct {
	Constants       map[string]string
	TempFiles       []string
	RootDir         string
	AllowedCommands []string
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

func (c *Cmd) Validate(s Status) (err error) {
	if len(c.Cmd) == 0 {
		err = errors.New("Command string must be non-empty.")
		return
	}

	cmd, err := exec.LookPath(c.Cmd[0])
	if err != nil {
		err = util.E.Annotate(err, "Command", c.Cmd[0], "could not be found")
		return
	}

	c.Cmd[0] = cmd
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

		command := splitWsQuote(line)
		if len(command) > 0 {
			cmd := &Cmd{command}
			err = cmd.Validate(Status{})
			if err != nil {
				return nil, err
			}
			c.Links = append(c.Links, cmd)
		}
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
