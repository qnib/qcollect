package test_utils

import (
	"strings"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBuildLogger(t *testing.T) {
	assert := assert.New(t)
	log := l.WithFields(l.Fields{"testing": true})
	assert.Equal(BuildLogger(), log)
}

// DirectoryOfCurrentFile returns directory of current file
// bit like __dir__ of Ruby
func TestDirectoryOfCurrentFile(t *testing.T) {
	assert := assert.New(t)
	res := strings.Split(DirectoryOfCurrentFile(), "/")
	for i := len(res)/2 - 1; i >= 0; i-- {
		ser := len(res) - 1 - i
		res[i], res[ser] = res[ser], res[i]
	}
	assert.Equal(res[0:4], []string{"test_utils", "qcollect", "qnib", "github.com"}, "Should be different")
}
