package util

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMakeRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "done")
	}))
	defer ts.Close()

	httpClient := new(HTTPAlive)
	httpClient.Configure(time.Duration(10)*time.Second, time.Minute, 10)
	assert.Equal(t, 10, httpClient.transport.MaxIdleConnsPerHost)

	httpClient.SetHeader(map[string]string{
		"foo": "bar",
	})

	assert.Equal(t, httpClient.customHeader["foo"], "bar")

	resp, err := httpClient.MakeRequest("GET", ts.URL, bytes.NewBufferString("qcollect"))

	assert.Nil(t, err)
	assert.Equal(t, string(resp.Body), "done\n")
	assert.Empty(t, httpClient.customHeader)
}
