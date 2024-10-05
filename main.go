package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

type Address string
type ifconfig struct {
	IPAddr string `json:"ip_addr"`
}

type Result struct {
	Addr Address
	RTT  time.Duration
}

func main() {
	addresses := flag.String("addresses", "", "Comma-separated list of IP addresses or FQDNs")
	pingCount := flag.Int("count", 3, "How many times to ping each address")
	flag.Parse()

	printIdentity()

	addrList := strings.Split(*addresses, ",")
	fmt.Printf("  Will ping each one of %s %d times\n", strings.Join(addrList, ", "), *pingCount)

	resultChan := make(chan Result)
	all := make(map[Address][]time.Duration)

	go func() {
		for {
			res := <-resultChan
			all[res.Addr] = append(all[res.Addr], res.RTT)
		}
	}()

	for _, address := range addrList {
		wg.Add(1)
		go ping(address, *pingCount, resultChan, &wg)
	}

	wg.Wait()
	printAverages(all)
}

func ping(address string, pingCount int, resChan chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	// Regular expression to extract the time from ping output (e.g., "time=20.1 ms")
	timeRegex := regexp.MustCompile(`time=([\d.]+) ms`)

	for i := 1; i <= pingCount; i++ {
		cmd := exec.Command("ping", "-c", "1", address)

		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to ping %s: %v\n", address, err)
			break
		}

		scanner := bufio.NewScanner(strings.NewReader(out.String()))
		for scanner.Scan() {
			line := scanner.Text()
			// Match the time in the ping output
			if matches := timeRegex.FindStringSubmatch(line); matches != nil {
				rtt, err := strconv.ParseFloat(matches[1], 64)
				if err == nil {
					resChan <- Result{Addr: Address(address), RTT: time.Duration(rtt) * time.Millisecond}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading ping output: %v\n", err)
		}
	}
}

func printAverages(all map[Address][]time.Duration) {
	type averageResult struct {
		Addr    Address
		Average time.Duration
	}

	averages := []averageResult{}

	for addr, durations := range all {
		var sum time.Duration
		for _, duration := range durations {
			sum += duration
		}
		average := sum / time.Duration(len(durations))
		averages = append(averages, averageResult{addr, average})
	}

	sort.Slice(averages, func(i, j int) bool {
		return averages[i].Average < averages[j].Average
	})

	slowest := make([]*averageResult, 2)
	for _, res := range averages {
		fmt.Printf("    - %s: %v\n", res.Addr, res.Average)
		slowest[1] = slowest[0]
		slowest[0] = &res
	}

	fmt.Printf("\n  Slowest: %s - %+v", slowest[0].Addr, slowest[0].Average)
	traceRoute(slowest[0].Addr)
	traceRoute(slowest[1].Addr)
}

func traceRoute(addr Address) {
	cmd := exec.Command("sh", "-c", "traceroute "+string(addr)+" | awk 'NR>1 {print $2}' | grep -v '*' | grep -v '^$' | uniq")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to traceroute %s: %v\n", addr, err)
		return
	}
	fmt.Printf("\n=========\n  Traceroute to %s:\n%s\n", addr, string(output))
}

func printIdentity() {

	// IPv4
	{

		// Custom Dialer to force IPv4
		dialer := &net.Dialer{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "tcp4", address)
				},
			},
		}

		// Custom Transport using the Dialer
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
			},
		}
		// Get external IP
		// Fetch IPv4 details
		resp4, err4 := client.Get("http://ifconfig.me/all.json")
		if err4 != nil {
			fmt.Printf("Failed to get IPv4 external IP: %v\n", err4)
		} else {
			defer resp4.Body.Close()
			var result4 ifconfig
			if err := json.NewDecoder(resp4.Body).Decode(&result4); err != nil {
				fmt.Printf("Failed to parse JSON for IPv4: %v\n", err)
			} else {
				// fmt.Printf("Fetched IPv4 details: %s\n", result4.IPAddr)
				getYou(result4.IPAddr)
			}
		}
	}

	// IPv6
	{
		// Custom Dialer to force IPv6
		dialer := &net.Dialer{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "tcp6", address)
				},
			},
		}

		// Custom Transport using the Dialer
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
			},
		}
		// Fetch IPv6 details
		resp6, err6 := client.Get("http://ifconfig.me/all.json")
		if err6 != nil {
			fmt.Printf("Failed to get IPv6 external IP: %v\n", err6)
		} else {
			defer resp6.Body.Close()
			var result6 ifconfig
			if err := json.NewDecoder(resp6.Body).Decode(&result6); err != nil {
				fmt.Printf("Failed to parse JSON for IPv6: %v\n", err)
			} else {
				// fmt.Printf("Fetched IPv6 details: %s\n", result6.IPAddr)
				getYou(result6.IPAddr)
			}
		}
	}

}

func getYou(ip string) {
	names, err := net.LookupAddr(ip)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("You are %s --> %s\n", names[0], ip)
}
