package util

import (
	"errors"
	"math"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/andygrunwald/megos"
	"github.com/stretchr/testify/assert"
)

const (
	testHosts    = "http://127.0.0.1:80/,http://some.other.ip:80/,http://some.external.ip:8080/"
	testHostsBad = "http://[fe80::%31%25en0]:8080/,http://[fe80::%31%25en0]/"
	testTTL      = 5 * time.Second
)

type MockSUT struct {
	MesosLeaderElect
	setCalled bool
}

func (sut *MockSUT) set() {
	sut.setCalled = true
	sut.leader = "testLeader"
}

func TestMesosLeaderElectConfigure(t *testing.T) {
	oldCreateMesos := createMesos
	defer func() { createMesos = oldCreateMesos }()

	called := false
	createMesos = func(u []*url.URL) *megos.Client {
		called = true
		return nil
	}

	mle := new(MesosLeaderElect)
	mle.Configure(testHosts, testTTL)

	assert.Equal(t, testTTL, mle.ttl, "TTL set is not the same as that passed via Configure func")
	assert.True(t, called, "Create new client was not called")
}

func TestMesosLeaderElectParseUrls(t *testing.T) {
	extract := func(urls []*url.URL) []string {
		ret := make([]string, len(urls), len(urls))
		for k, v := range urls {
			ret[k] = v.String()
		}

		return ret
	}

	mle := new(MesosLeaderElect)

	tests := []struct {
		testUrls    string
		doesSucceed bool
		msg         string
	}{
		{testHosts, true, "Valid hosts string; should pass."},
		{testHostsBad, false, "Valid hosts string; should pass."},
	}

	for _, test := range tests {
		switch test.doesSucceed {
		case true:
			assert.Equal(t, strings.Split(test.testUrls, ","), extract(mle.parseUrls(test.testUrls)), test.msg)
		case false:
			assert.Empty(t, mle.parseUrls(test.testUrls), test.msg)
		}
	}
}

func TestMesosLeaderElectGet(t *testing.T) {
	var oldDetermineLeader = determineLeader
	defer func() { determineLeader = oldDetermineLeader }()

	setCalled := false
	determineLeader = func(c *megos.Client) (*megos.Pid, error) {
		setCalled = true
		ret := megos.Pid{"", "testLeader", 8080}

		return &ret, nil
	}

	var tests = []struct {
		initialLeader string
		expireAfter   time.Duration
		ttl           time.Duration
		isSetCalled   bool
		expectedRet   string
		explanation   string
	}{
		{"", 5 * time.Minute, 15 * time.Minute, true, "testLeader", "No leader, set should be called"},
		{"testLeader2", -5 * time.Minute, 4 * time.Minute, true, "testLeader", "Leader present but expired, set should be called"},
		{"testLeader2", 5 * time.Minute, 6 * time.Minute, false, "testLeader2", "Leader present and not expired, set should not be called"},
	}

	for _, test := range tests {
		sut := MesosLeaderElect{test.initialLeader, nil, test.ttl, time.Now().Add(test.expireAfter)}
		setCalled = false

		ret := sut.Get()

		assert.Equal(t, test.expectedRet, ret)
		assert.Equal(t, test.isSetCalled, setCalled, test.explanation)
	}
}

func TestMesosLeaderElectSet(t *testing.T) {
	var oldDetermineLeader = determineLeader
	defer func() { determineLeader = oldDetermineLeader }()

	tests := []struct {
		err          error
		expectLeader bool
		msg          string
	}{
		{nil, true, "Leader returned without error, all should be ok."},
		{errors.New("Test error"), false, "Leader returned without error, all should be ok."},
	}

	for _, test := range tests {
		determineLeader = func(c *megos.Client) (*megos.Pid, error) {
			ret := megos.Pid{"", "testLeader", 8080}
			return &ret, test.err
		}

		sut := new(MesosLeaderElect)
		sut.Configure("http://1.2.3.4/", 5*time.Second)

		assert.Equal(t, true, time.Time.IsZero(sut.expire), "When the struct is just initialized, expire should not be set to anything.")

		sut.set()

		assert.Equal(t, math.Ceil(testTTL.Seconds()), math.Ceil(sut.expire.Sub(time.Now()).Seconds()), "After the struct's set method's called, expire should be set to now + ttl.")

		switch test.expectLeader {
		case true:
			assert.Equal(t, "testLeader", sut.leader, test.msg)
		case false:
			assert.Empty(t, sut.leader, test.msg)
		}
	}
}
