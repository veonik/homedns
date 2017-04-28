# homedns

homedns is a cross-platform utility to update a Linode managed DNS A record
with the system's public IP address.

# Installation

[Download the latest release](https://github.com/veonik/homedns/releases/latest) 
as a pre-built binary for your platform.

### Using go get

Install with `go get`.

```bash
go get github.com/veonik/homedns
```

# Usage

```
Usage of homedns:
  -domain string
    	DNS Domain name, required
  -key string
    	Linode API key, required
  -name string
    	DNS A Record name, required
  -verbose
    	Enable verbose logging
  -help
    	Show this help text
```
