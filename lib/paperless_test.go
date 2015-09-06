package paperless

import (
	"os"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestChecksumSame(t *testing.T) {
	first := "First string"
	second := "Second string"
	sum1 := Checksum([]byte(first))
	sum2 := Checksum([]byte(second))

	assert.NotEmpty(t, sum1)
	assert.NotEqual(t, sum1, sum2)
}

func TestChecksumFileSame(t *testing.T) {
	first := "First string"
	sum1 := Checksum([]byte(first))
	var sum2 string

	fp, err := ioutil.TempFile("", "testfile")
	assert.Nil(t, err)
	defer os.Remove(fp.Name())
	defer fp.Close()
	fp.Write([]byte(first))
	sum2, err = ChecksumFile(fp.Name())
	assert.Nil(t, err)
	assert.Equal(t, sum1, sum2)
}
