package util

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// HTTPAlive implements a simple way of reusing http connections
type HTTPAlive struct {
	client       *http.Client
	transport    *http.Transport
	customHeader map[string]string
}

// HTTPAliveResponse returns a response
type HTTPAliveResponse struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

// Configure the http connection
func (connection *HTTPAlive) Configure(timeout time.Duration,
	aliveDuration time.Duration,
	maxIdleConnections int) {
	if connection.transport == nil {
		connection.transport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: aliveDuration,
			}).Dial,
			MaxIdleConnsPerHost: maxIdleConnections,
		}
	}

	if connection.client == nil {
		connection.client = &http.Client{
			Transport: connection.transport,
		}
	}
}

// SetHeader for setting some custom headers
func (connection *HTTPAlive) SetHeader(header map[string]string) {
	connection.customHeader = header
}

// MakeRequest make a new http request
func (connection *HTTPAlive) MakeRequest(method string,
	uri string, body io.Reader) (*HTTPAliveResponse, error) {

	defer connection.resetCustomHeader()
	req, err := http.NewRequest(method, uri, body)

	if err != nil {
		return nil, err
	}

	// Apply user provided headers
	for key, value := range connection.customHeader {
		req.Header.Set(key, value)
	}

	return connection.submitRequest(req)
}

func (connection *HTTPAlive) submitRequest(req *http.Request) (*HTTPAliveResponse, error) {
	rsp, err := connection.client.Do(req)

	if rsp != nil {
		defer discardResponseBody(rsp.Body)
	}

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	httpAliveResponse := new(HTTPAliveResponse)
	httpAliveResponse.Body = body
	httpAliveResponse.StatusCode = rsp.StatusCode
	httpAliveResponse.Header = rsp.Header
	return httpAliveResponse, nil
}

func (connection *HTTPAlive) resetCustomHeader() {
	connection.customHeader = make(map[string]string)
}

func discardResponseBody(body io.ReadCloser) {
	io.Copy(ioutil.Discard, body)
	body.Close()
}
