# DNS Resolver

A simple DNS resolver implemented in Go that demonstrates how DNS resolution works from the ground up. This program performs recursive DNS resolution starting from a root nameserver, similar to how a DNS resolver works in practice.

## Features

- Performs recursive DNS resolution
- Supports A (address) and NS (nameserver) record lookups
- Handles DNS message encoding and decoding
- Follows the DNS protocol specification
- Starts from root nameserver (198.41.0.4) and traverses the DNS hierarchy

## Prerequisites

- Go 1.15 or higher
- Basic understanding of DNS concepts

## Installation

1. Clone the repository:
```bash
git clone https://github.com/biswaone/dns-go
cd dns-go
```

2. Build the program:
```bash
go build
```

## Usage

Run the program with a domain name and record type as arguments:

```bash
./dns-go <record_type> <domain_name> 
```

Where:
- `<domain_name>` is the domain you want to look up (e.g., example.com)
- `<record_type>` is either "A" or "NS" (case-insensitive)

### Examples

Look up an A record (IP address):
```bash
./dns-go A example.com 
```

Look up an NS record (nameserver):
```bash
./dns-go NS example.com 
```

### Sample Output

For A record lookup:
```
Querying 198.41.0.4 for example.com
Querying 192.41.162.30 for example.com
Querying 198.41.0.4 for a.iana-servers.net
Querying 192.55.83.30 for a.iana-servers.net
Querying 199.43.135.53 for a.iana-servers.net
Querying 199.43.135.53 for example.com
The IP of example.com is 23.192.228.80
```

For NS record lookup:
```
Querying 198.41.0.4 for example.com
Querying 192.41.162.30 for example.com
Querying 198.41.0.4 for a.iana-servers.net
Querying 192.55.83.30 for a.iana-servers.net
Querying 199.43.135.53 for a.iana-servers.net
Querying 199.43.135.53 for example.com
The nameserver for example.com is a.iana-servers.net
```

## How It Works

1. The program starts with a root nameserver (198.41.0.4)
2. It sends a DNS query for the requested domain and record type
3. If the nameserver doesn't have the answer, it returns:
   - Either the IP of a more specific nameserver
   - Or the domain name of a more specific nameserver
4. The program follows these referrals until it gets the final answer
5. For NS queries, it returns the authoritative nameserver
6. For A queries, it returns the IP address

## Acknowledgments

- Based on Julia Evans' [Implement DNS in a weekend](https://implement-dns.wizardzines.com/) 
