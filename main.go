package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

type Address string
type Result struct {
	Addr Address
	Time time.Duration
}

func main() {
	// Define a string slice to hold IP addresses or FQDNs
	addresses := flag.String("addresses", "", "Comma-separated list of IP addresses or FQDNs")
	pingCount := flag.Int("count", 3, "How many times to ping each address")

	// Parse the command-line flags
	flag.Parse()

	// Split the addresses by comma
	addrList := strings.Split(*addresses, ",")
	if addrList == nil || len(addrList) <= 0 || addrList[0] == "" {
		fmt.Printf("--addresses is empty\n")
		return
	}
	fmt.Printf("Will ping %+v %d times\n", addrList, *pingCount)

	resultChan := make(chan Result)
	all := make(map[Address][]time.Duration)

	go func() {
		for {
			res, ok := <-resultChan
			if !ok {
				// fmt.Printf(">>> done <<<\n")
				break
			}
			// fmt.Printf("Received result: %+v\n", res)
			addr := res.Addr
			all[addr] = append(all[addr], res.Time)
		}
	}()

	// Print each address
	for _, address := range addrList {
		wg.Add(1)
		go ping(address, *pingCount, resultChan, &wg)
	}

	wg.Wait()

	printAverages(all)

	// fmt.Printf("Here is the collection: %+v\n", all)
}

func ping(address string, pingCount int, resChan chan<- Result, wg *sync.WaitGroup) {
	counter := 1
	for counter <= pingCount {
		fmt.Printf("%d: Pinging %s\n", counter, address)
		start := time.Now()
		cmd := exec.Command("ping", "-c 1", address)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Failed to ping %s: %v\n", address, err)
			resChan <- Result{Address(address), -1}
		} else {
			duration := time.Since(start)
			resChan <- Result{Address(address), duration}
		}
		counter += 1
	}
	wg.Done()
}

func printAverages(all map[Address][]time.Duration) {
	for addr, durations := range all {
		var sum time.Duration
		for _, duration := range durations {
			sum += duration
		}
		average := sum / time.Duration(len(durations))
		fmt.Printf("Address: %s Average Duration: %v\n", addr, average)
	}
}
