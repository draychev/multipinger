package main

import (
	"flag"
	"fmt"
)

func main() {
	// Define a string slice to hold IP addresses or FQDNs
	addresses := flag.String("addresses", "", "Comma-separated list of IP addresses or FQDNs")

	// Parse the command-line flags
	flag.Parse()

	// Split the addresses by comma
	addrList := strings.Split(*addresses, ",")

	// Print each address
	for _, address := range addrList {
		fmt.Println(address)
	}
}