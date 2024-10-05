# multipinger

This tool pings a list of IP addresses or FQDNs and returns a list of the average ping times.
Then it traceroutes the slowest ones.

Why is this useful?
It is mainly fun to run this as you travel with your mifi/starlink and watch your latency and routes to a list of interesting destinations change as you move.

Usage:
```bash
go run ./main.go --addresses="8.8.8.8,mirrors.fcix.net"
```
