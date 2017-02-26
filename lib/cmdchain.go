
package paperless

type Environment struct {
	Constants map[string]string
	TempFiles []string
	RootDir string
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

