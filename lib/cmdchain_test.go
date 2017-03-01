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
			Links:[]Link{&Cmd{[]string{"/bin/true"}}},
		}, false},

		{"Two commands", args{"true\nfalse"}, &CmdChain{
			Links:[]Link{
				&Cmd{[]string{"/bin/true"}},
				&Cmd{[]string{"/bin/false"}},
			},
		}, false},

		{"Arguments", args{"true first second"}, &CmdChain{
			Links:[]Link{
				&Cmd{[]string{"/bin/true", "first", "second"}},
			},
		}, false},

		{"Quoted arguments", args{"true 'first second'"}, &CmdChain{
			Links:[]Link{
				&Cmd{[]string{"/bin/true", "first second"}},
			},
		}, false},

		{"Command not found", args{"this-command-is-not-found"}, nil, true},
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
