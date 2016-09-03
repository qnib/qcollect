package util

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileSizeMissing(t *testing.T) {
	size, err := GetFileSize("/random/file")
	assert.Equal(t, int64(0), size)
	assert.NotNil(t, err)
}

func TestGetFileSize(t *testing.T) {
	file, _ := ioutil.TempFile("", "my.cnf")
	defer os.Remove(file.Name())

	file.Write([]byte("abcdef"))

	size, err := GetFileSize(file.Name())
	assert.Equal(t, int64(6), size)
	assert.Nil(t, err)
}
