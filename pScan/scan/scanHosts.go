// package scan provides types and functions to perform TCP port scans on a list of hosts
package scan

import (
	"fmt"
	"net"
	"time"
)

// PortState represents the state of a single TCP port
type PortState struct {
	Port int
	Open state
}

type state bool

// String converts the boolean value of state to a human readable string
// By implementing the String method on the state type, you satisfy the Stringer interface, which allows you to use this type directly with print functions
func (s state) String() string {
	if s {
		return "open"
	}

	return "closed"
}

// scanPort performs a port scan on a single TCP port
func scanPort(host string, port int, timeout time.Duration) PortState {
	// boolean state automatically set to false (zero value)
	p := PortState {
		Port: port,
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	scanConn, err := net.DialTimeout("tcp", address, time.Duration(timeout) * time.Second)

	// couldn't connect, and returned error
	if err != nil {
		return p
	}

	// if connection succeeds
	scanConn.Close()
	p.Open = true
	return p
}	


// Result represents the scan results for a single host
type Results struct {
	Host string
	NotFound bool
	PortStates []PortState
}

func Run(hl *HostsList, ports []int, timeout time.Duration) []Results {
	res := make([]Results, 0, len(hl.Hosts))

	for _, h := range hl.Hosts {
		r := Results {
			Host: h,
		}

		if _, err := net.LookupHost(h); err != nil {
			r.NotFound = true
			res = append(res, r)
			continue
		}

		for _, p := range ports {
			r.PortStates = append(r.PortStates, scanPort(h, p, timeout))
		}

		res = append(res, r)
	}

	return res
}