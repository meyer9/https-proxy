# https-proxy

`https-proxy` is a simple tool to proxy a local http server to a https.

## Installation

```bash
go install github.com/meyer9/https-proxy
```

## Usage

```bash
Usage of https-proxy:
  -addr string
    address to listen on for https server (default "127.0.0.1:1443")
  -cert string
    certificate file (default "./localhost.pem")
  -key string
    certificate key file (default "./localhost-key.pem")
  -proxy string
    address to proxy requests to (default "127.0.0.1:3000")
```

## Example

```bash
mkcert localhost
https-proxy -proxy "127.0.0.1:3000"
```

Then, you can navigate to <https://localhost:1443/> to access the proxied application.
