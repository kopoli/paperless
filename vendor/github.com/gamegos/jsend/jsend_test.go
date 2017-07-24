package jsend

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestWrap(t *testing.T) {
	w := httptest.NewRecorder()
	jw := Wrap(w)
	jw2 := Wrap(jw)

	if jw2 != jw {
		t.Errorf("Wrap: have: %q, want: %q", jw2, jw)
	}
}

func TestHeader(t *testing.T) {
	w := httptest.NewRecorder()
	jw := Wrap(w)

	if !reflect.DeepEqual(jw.Header(), w.Header()) {
		t.Errorf("Header: have: %q, want: %q", jw.Header(), w.Header())
	}
}

var (
	dataNil   interface{}
	dataEmpty = []byte(nil)
	data1     = map[string]interface{}{"foo": "bar", "baz": "qux"}
	data2     = map[string]interface{}{"id": "invalid", "dob": "empty"}
	// cant json encode this
	dataInvalid = map[string]interface{}{"foos": map[[2]byte]string{
		[2]byte{'2', '3'}: "4",
		[2]byte{'a', 'b'}: "c",
	}}
)

type caseIn struct {
	code    int
	data    interface{}
	message string
}

type caseOut struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

var testCases = []struct {
	in  *caseIn
	out *caseOut
}{
	{&caseIn{200, data1, ""}, &caseOut{StatusSuccess, data1, ""}},
	{&caseIn{0, data1, ""}, &caseOut{StatusSuccess, data1, ""}},
	{&caseIn{200, dataNil, ""}, &caseOut{StatusSuccess, dataNil, ""}},

	{&caseIn{400, data2, ""}, &caseOut{StatusFail, data2, ""}},

	{&caseIn{500, dataNil, "foo"}, &caseOut{StatusError, dataNil, "foo"}},
	{&caseIn{503, dataNil, ""}, &caseOut{StatusError, dataNil, "Service Unavailable"}},
}

func TestSend(t *testing.T) {
	for _, tt := range testCases {
		w := httptest.NewRecorder()
		jw := Wrap(w)

		if tt.in.data != dataNil {
			jw.Data(tt.in.data)
		}

		if tt.in.code != 0 {
			jw.Status(tt.in.code)
		}

		if tt.in.message != "" {
			jw.Message(tt.in.message)
		}
		jw.Send()

		var body *caseOut
		json.Unmarshal(w.Body.Bytes(), &body)

		if !reflect.DeepEqual(body, tt.out) {
			t.Errorf("Send: have: %#v, want: %#v", body, tt.out)
		}
	}
}

func TestSentAlready(t *testing.T) {
	jw := Wrap(httptest.NewRecorder())

	n, err := jw.Data("xx").Send()

	if n == 0 || err != nil {
		t.Errorf("SentAlready: have: (%d, %q), want: (>0, nil)", n, err)
	}

	n2, err2 := jw.Send()

	if n2 != 0 || err2 != errSentAlready {
		t.Errorf("SentAlready: have: (%d, %q), want: (0, %q)", n2, err2, errSentAlready)
	}
}

func TestJSONErr(t *testing.T) {
	jw := Wrap(httptest.NewRecorder())

	n, err := jw.Data(dataInvalid).Send()

	if n != 0 || err == nil {
		t.Errorf("Send invalid json: have: (%d, %q), want: (>0, json: unsupported type: error)", n, err)
	}
}
