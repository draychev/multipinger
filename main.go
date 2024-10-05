package main

import (
	"flag"
	"fmt"
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
				fmt.Printf(">>> done <<<\n")
				break
			}
			fmt.Printf("Received result: %+v\n", res)
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

	fmt.Printf("Here is the collection: %+v\n", all)
}

func ping(address string, pingCount int, resChan chan<- Result, wg *sync.WaitGroup) {
	counter := 1
	for counter <= pingCount {
		fmt.Printf("%d: Pinging %s\n", counter, address)
		resChan <- Result{Address(address), 789}
		counter += 1
	}
	wg.Done()
}
