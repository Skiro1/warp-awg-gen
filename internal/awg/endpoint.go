package awg

import (
	"fmt"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var knownWarpIPs = []string{
	"162.159.192.11",
	"162.159.192.21",
	"162.159.192.31",
	"162.159.192.41",
	"162.159.192.51",
	"162.159.192.61",
	"162.159.192.71",
	"162.159.192.81",
	"162.159.192.91",
	"162.159.192.14",
	"162.159.192.24",
	"162.159.192.34",
	"162.159.192.44",
	"162.159.192.54",
	"162.159.192.64",
	"162.159.192.74",
	"162.159.192.84",
	"162.159.192.94",
	"162.159.192.101",
	"162.159.192.105",
	"162.159.192.106",
	"162.159.192.205",
	"162.159.192.215",
	"162.159.192.1",
}

var warpPorts = []int{
	854, 859, 864, 878, 880, 890, 891, 894, 903, 908,
	928, 934, 939, 942, 943, 945, 946, 955, 968, 987,
	988, 1002, 1010, 1014, 1018, 1070, 1074, 1180, 1387, 1843,
	2371, 2506, 3138, 3476, 3581, 3854, 4177, 4198, 4233, 5279,
	5956, 7103, 7152, 7156, 7281, 7559, 8319, 8742, 8854, 8886,
}

type probeResult struct {
	ip    string
	rtt   time.Duration
	reach bool
}

func SelectFastestEndpoint() (string, time.Duration) {
	addrs := resolveWarpIPs()

	results := probeIPs(addrs)

	var reachable []probeResult
	for _, r := range results {
		if r.reach {
			reachable = append(reachable, r)
		}
	}

	if len(reachable) == 0 {
		port := warpPorts[rng.Intn(len(warpPorts))]
		return fmt.Sprintf("162.159.192.1:%d", port), 0
	}

	sort.Slice(reachable, func(i, j int) bool {
		return reachable[i].rtt < reachable[j].rtt
	})

	port := warpPorts[rng.Intn(len(warpPorts))]
	return fmt.Sprintf("%s:%d", reachable[0].ip, port), reachable[0].rtt
}

func resolveWarpIPs() []string {
	seen := make(map[string]bool)
	var addrs []string

	hosts := []string{
		"engage.cloudflareclient.com",
		"mfi.tribukvy.ltd",
		"mpl.tribukvy.ltd",
	}

	for _, host := range hosts {
		ips, err := net.LookupHost(host)
		if err != nil {
			continue
		}
		for _, ip := range ips {
			if !seen[ip] {
				seen[ip] = true
				addrs = append(addrs, ip)
			}
		}
	}

	for _, ip := range knownWarpIPs {
		if !seen[ip] {
			seen[ip] = true
			addrs = append(addrs, ip)
		}
	}

	return addrs
}

func probeIPs(addrs []string) []probeResult {
	var mu sync.Mutex
	results := make([]probeResult, 0, len(addrs))
	var wg sync.WaitGroup

	for _, ip := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			start := time.Now()
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(addr, "443"), 2*time.Second)
			rtt := time.Since(start)
			r := probeResult{ip: addr, rtt: rtt, reach: err == nil}
			if conn != nil {
				conn.Close()
			}
			mu.Lock()
			results = append(results, r)
			mu.Unlock()
		}(ip)
	}

	wg.Wait()
	return results
}
