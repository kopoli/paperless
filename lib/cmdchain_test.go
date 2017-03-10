package paperless

import (
	"reflect"
	"testing"
)

func Test_parseConsts(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"No variables", args{"something else"}, []string{}},
		{"Single variable", args{"$a"}, []string{"a"}},
		{"Many variables", args{"$a $b $c"}, []string{"a", "b", "c"}},
		{"Combined", args{"$first$second"}, []string{"first", "second"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseConsts(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseConsts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCmdChainScript(t *testing.T) {
	type args struct {
		script string
	}
	tests := []struct {
		name    string
		args    args
		wantC   *CmdChain
		wantErr bool
	}{
		{"Empty", args{""}, &CmdChain{}, false},

		{"Comment and empty line", args{`
# comment`}, &CmdChain{}, false},

		{"Single command", args{"true"}, &CmdChain{
			Links: []Link{&Cmd{[]string{"true"}}},
		}, false},

		{"Two commands", args{"true\nfalse"}, &CmdChain{
			Links: []Link{
				&Cmd{[]string{"true"}},
				&Cmd{[]string{"false"}},
			},
		}, false},

		{"Arguments", args{"true first second"}, &CmdChain{
			Links: []Link{
				&Cmd{[]string{"true", "first", "second"}},
			},
		}, false},

		{"Quoted arguments", args{"true 'first second'"}, &CmdChain{
			Links: []Link{
				&Cmd{[]string{"true", "first second"}},
			},
		}, false},

		{"Included a constant", args{"true $variable"}, &CmdChain{
			Environment: Environment{
				Constants: map[string]string{
					"variable": "",
				},
			},
			Links: []Link{
				&Cmd{[]string{"true", "$variable"}},
			},
		}, false},

		{"Command not found", args{"this-command-is-not-found"}, nil, true},

		{"Included a temporary file", args{"true $tmpSomething"}, &CmdChain{
			Environment: Environment{
				Constants: map[string]string{
					"tmpSomething": "",
				},
				TempFiles: []string{"tmpSomething"},
			},
			Links: []Link{
				&Cmd{[]string{"true", "$tmpSomething"}},
			},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, err := NewCmdChainScript(tt.args.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCmdChainScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("NewCmdChainScript() = %v, want %v", gotC, tt.wantC)
			}
		})
	}
}

func Test_splitWsQuote(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Empty", args{""}, []string{}},
		{"One item", args{"jep"}, []string{"jep"}},
		{"Two items", args{"jep something"}, []string{"jep", "something"}},
		{"Quoted", args{"'sth abc'"}, []string{"sth abc"}},
		{"Quoted two", args{"'sth' abc"}, []string{"sth", "abc"}},
		{"Mixed quotes", args{"'a b c' \"c  e\""}, []string{"a b c", "c  e"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitWsQuote(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitWsQuote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCmd_Validate(t *testing.T) {
	type fields struct {
		Cmd []string
	}
	type args struct {
		e Environment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Proper command", fields{[]string{"true"}},
			args{Environment{
				RootDir: "/",
			}}, false},
		{"Empty command", fields{[]string{""}}, args{}, true},
		// {"LastErr already set", fields{[]string{"true"}}, args{Status{LastErr: errors.New("abc")}}, true},
		{"Command not found", fields{[]string{"command-is-not-found"}}, args{}, true},
		{"Command is allowed", fields{[]string{"true"}},
			args{Environment{
				RootDir: "/",
				AllowedCommands: map[string]bool{
					"true": true,
				},
			}}, false},
		{"Command not allowed", fields{[]string{"true"}},
			args{Environment{
				AllowedCommands: map[string]bool{
					"b": true,
				},
			}}, true},
		{"Constant is defined", fields{[]string{"true", "$something"}},
			args{Environment{
				RootDir: "/",
				Constants: map[string]string{
					"something": "value",
				},
			}}, false},

		{"Constant is not defined", fields{[]string{"true", "$else"}},
			args{Environment{
					RootDir: "/",
				}}, true},

		{"Commands cannot be read from a constant", fields{[]string{"$cmd"}},
			args{Environment{
					Constants: map[string]string{
						"cmd": "true",
					},
				}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cmd{
				Cmd: tt.fields.Cmd,
			}
			if err := c.Validate(&tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("Cmd.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
