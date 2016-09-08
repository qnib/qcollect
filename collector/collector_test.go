package collector

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	names := []string{"Test", "Diamond"}
	for _, name := range names {
		c := New(name)
		name = strings.Split(name, " ")[0]
		assert.NotNil(t, c, "should create a Collector for "+name)
		assert.Equal(t, name, c.Name())
		assert.Equal(
			t,
			DefaultCollectionInterval,
			c.Interval(),
			"should be the default collection interval for "+name,
		)
		assert.Equal(
			t,
			name+"Collector",
			fmt.Sprintf("%s", c),
			"String() should append Collector to the name for "+name,
		)

		c.SetInterval(999)
		assert.Equal(t, 999, c.Interval(), "should have set the interval")
	}
}

func TestNewInvalidCollector(t *testing.T) {
	c := New("INVALID COLLECTOR")
	assert.Nil(t, c, "should not create a Collector")
}
