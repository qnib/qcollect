package test_utils

import (
	"path"
	"runtime"

	l "github.com/Sirupsen/logrus"
)

func BuildLogger() *l.Entry {
	return l.WithFields(l.Fields{"testing": true})
}

// DirectoryOfCurrentFile returns directory of current file
// bit like __dir__ of Ruby
func DirectoryOfCurrentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
