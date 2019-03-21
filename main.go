package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
	proxyTo string
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Scheme == "http" {
		r.URL.Scheme = "https"
		http.Redirect(w, r, r.URL.String(), 301)
	}

	client := &http.Client{}

	newReq, err := http.NewRequest(r.Method, fmt.Sprintf("http://%s%s", p.proxyTo, r.RequestURI), r.Body)
	if err != nil {
		panic(err)
	}

	delHopHeaders(r.Header)

	newReq.Header = r.Header

	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		appendHostToXForwardHeader(newReq.Header, clientIP)
	}

	resp, err := client.Do(newReq)
	if err != nil {
		http.Error(w, "Server Error", http.StatusBadGateway)
		log.Printf("error: %s", err)
	}

	defer resp.Body.Close()

	delHopHeaders(resp.Header)

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:1443", "address to listen on for https server")
	var proxied = flag.String("proxy", "127.0.0.1:3000", "address to proxy requests to")
	var cert = flag.String("cert", "./localhost.pem", "certificate file")
	var key = flag.String("key", "./localhost-key.pem", "certificate key file")

	flag.Parse()

	_, err := os.Stat(*cert)
	if err != nil {
		panic(err)
	}

	_, err = os.Stat(*key)
	if err != nil {
		panic(err)
	}

	handler := &proxy{
		proxyTo: *proxied,
	}

	log.Printf("Starting https server on %s proxying to %s", *addr, *proxied)

	err = http.ListenAndServeTLS(*addr, *cert, *key, handler)
	if err != nil {
		panic(err)
	}
}
