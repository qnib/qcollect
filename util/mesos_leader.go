package util

import (
	"net/url"
	"strings"
	"time"

	l "github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
)

// Dependency injection: Makes writing unit tests much easier, by being able to override these values in the *_test.go files.
var (
	createMesos     = megos.NewClient
	determineLeader = (*megos.Client).DetermineLeader
	defaultLog      = l.WithFields(l.Fields{"app": "qcollect", "pkg": "util"})
)

// MesosLeaderElectInterface Interface to allow injecting MesosLeaderElect, for easier testing
type MesosLeaderElectInterface interface {
	Configure(string, time.Duration)
	Get() string
	set()
}

// MesosLeaderElect Encapsulation for the mesos leader given a set of masters. Cache this so that we don't spend too much time determining the leader every *collector.Collect() call, which happens every 10 seconds.
type MesosLeaderElect struct {
	leader string
	mesos  *megos.Client
	ttl    time.Duration
	expire time.Time
}

// Configure Provide the set of masters like so ("http://1.2.3.4:5050/,http://5.6.7.8:5050/") and the desired TTL for the cache.
func (mle *MesosLeaderElect) Configure(nodes string, ttl time.Duration) {
	hosts := mle.parseUrls(nodes)

	mle.ttl = ttl
	mle.mesos = createMesos(hosts)
}

// Get get the IP of the leader; calls *MesosLeaderElect.set() on the first call or if the TTL has expired.
func (mle *MesosLeaderElect) Get() string {
	if len(mle.leader) == 0 || time.Now().After(mle.expire) {
		mle.set()
	}

	return mle.leader
}

// parseUrls Conver the provided string of masters ("http://1.2.3.4:5050/,http://5.6.7.8:5050/") via *MesosLeaderElect.Configure() into an array of url.URLs, which is understood by the megos package.
func (mle *MesosLeaderElect) parseUrls(nodes string) []*url.URL {
	n := strings.Split(nodes, ",")
	hosts := make([]*url.URL, 0, len(n))

	for i, el := range n {
		if nodeURL, err := url.Parse(el); err == nil {
			hosts = hosts[0 : i+1]
			hosts[i] = nodeURL
		} else {
			defaultLog.Error("URL specified (", el, ") is invalid and cannot be parsed: ", err.Error())
		}
	}

	return hosts
}

// set Calls megos.client.DetermineLeader.
func (mle *MesosLeaderElect) set() {
	defer func() { mle.expire = time.Now().Add(mle.ttl) }()

	if leader, err := determineLeader(mle.mesos); err != nil {
		defaultLog.Error("Unable to determine mesos leader", err.Error())
	} else {
		mle.leader = leader.Host
	}
}
