package util

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// overriddenInterfaces Utility fixture to be used across tests—overrides net.Interfaces.
func overriddenInterfaces(flag net.Flags) func() ([]net.Interface, error) {
	return func() ([]net.Interface, error) {
		hw, _ := net.ParseMAC("a0:99:9b:13:f5:f1")
		ifaces := net.Interface{
			Index:        1,
			MTU:          1200,
			Name:         "test-0",
			HardwareAddr: hw,
			Flags:        flag,
		}

		return []net.Interface{ifaces}, nil
	}
}

//// overriddenInterfaces Utility fixture to be used across tests—overrides (*net.Interface).Addrs
func overriddenGetAddrs(addr net.Addr, err error) func(*net.Interface) ([]net.Addr, error) {
	if addr == nil {
		return func(i *net.Interface) ([]net.Addr, error) {
			return nil, err
		}
	}
	return func(i *net.Interface) ([]net.Addr, error) {
		return []net.Addr{addr}, nil
	}
}

func TestExternalIPErrorOnNetInterfaces(t *testing.T) {
	oldInterfaces := interfaces
	defer func() { interfaces = oldInterfaces }()

	expecterErr := errors.New("Test error")
	interfaces = func() ([]net.Interface, error) {
		return nil, expecterErr
	}

	ip, actualErr := ExternalIP()

	assert.Equal(t, "", ip)
	assert.Equal(t, expecterErr, actualErr)
}

func TestExternalIPEmptyIfaces(t *testing.T) {
	oldInterfaces := interfaces
	defer func() { interfaces = oldInterfaces }()

	interfaces = func() ([]net.Interface, error) {
		return []net.Interface{}, nil
	}

	ip, actualErr := ExternalIP()

	assert.Equal(t, "", ip)
	assert.Equal(t, "are you connected to the network?", actualErr.Error())
}

func TestExternalIPGetAddrFails(t *testing.T) {
	oldInterfaces := interfaces
	oldGetAddrs := getAddrs
	defer func() { interfaces = oldInterfaces; getAddrs = oldGetAddrs }()

	expectedErr := errors.New("Test error")
	interfaces = overriddenInterfaces(net.FlagUp | net.FlagMulticast)
	getAddrs = overriddenGetAddrs(nil, expectedErr)

	ip, actualErr := ExternalIP()

	assert.Equal(t, expectedErr, actualErr)
	assert.Empty(t, ip)
}

func TestExternalIPIsLoopback(t *testing.T) {
	oldInterfaces := interfaces
	oldGetAddrs := getAddrs
	defer func() { interfaces = oldInterfaces; getAddrs = oldGetAddrs }()

	_, ipn, _ := net.ParseCIDR("127.0.0.1/32")
	interfaces = overriddenInterfaces(net.FlagUp | net.FlagMulticast)
	getAddrs = overriddenGetAddrs(ipn, nil)

	ip, err := ExternalIP()

	assert.Empty(t, ip)
	assert.Equal(t, "are you connected to the network?", err.Error())
}

func TestExternalIPOnFlags(t *testing.T) {
	oldInterfaces := interfaces
	oldGetAddrs := getAddrs
	defer func() { interfaces = oldInterfaces; getAddrs = oldGetAddrs }()

	tests := []struct {
		flag      net.Flags
		ipPresent bool
	}{
		{0x0, false},
		{net.FlagUp | net.FlagMulticast, true},
		{net.FlagLoopback | net.FlagUp, false},
	}

	for _, test := range tests {
		_, ipn, _ := net.ParseCIDR("1.2.3.4/32")
		interfaces = overriddenInterfaces(test.flag)
		getAddrs = overriddenGetAddrs(ipn, nil)

		ip, err := ExternalIP()

		switch test.ipPresent {
		case true:
			assert.Equal(t, "1.2.3.4", ip)
			assert.Empty(t, err)
		case false:
			assert.Empty(t, ip)
			assert.Equal(t, errors.New("are you connected to the network?"), err)
		}

	}
}
