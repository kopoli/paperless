package paperless

import (
	"regexp"
	"strings"
	"unicode"
)

type Environment struct {
	Constants map[string]string
	TempFiles []string
	RootDir   string
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
	Cmd string
}

func (c *Cmd) Validate(Status) error {
	panic("not implemented")
}

func (c *Cmd) Run(Status) error {
	panic("not implemented")
}

////////////////////////////////////////////////////////////

var (
	constRe = regexp.MustCompile("\\$(\\w+)")
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
			return true;
		default:
			return unicode.IsSpace(r)
		}
	})
}
