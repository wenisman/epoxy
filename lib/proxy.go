package lib

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// write the HTTP connection message
func writeTCPConnection(conn *net.TCPConn, req *http.Request) (int, error) {
	var lines []string
	lines = append(
		lines,
		fmt.Sprintf("%s %s %s", req.Method, req.Host, req.Proto),
		fmt.Sprintf("Host: %s", req.Host))

	for k, v := range req.Header {
		for _, vv := range v {
			lines = append(lines, fmt.Sprintf("%s: %s", k, vv))
		}
	}

	http := fmt.Sprintf("%s\r\n\r\n", strings.Join(lines, "\r\n"))

	return conn.Write([]byte(http))
}

// read the TCP connection until the EOF is received
func readTCPResponse(conn *net.TCPConn) ([]byte, error) {
	buf := &bytes.Buffer{}

	for {
		data := make([]byte, 256)
		n, err := conn.Read(data)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}

			break
		}

		buf.Write(data[:n])
		if bytes.HasSuffix(data[:n], []byte("\r\n\r\n")) {
			break
		}
	}

	return buf.Bytes(), nil
}

// use tcp tunneling as for CONNECT as this covers https and ws
func handleTunneling(w http.ResponseWriter, r *http.Request) {
	if isAllowedEndpoint(r.Host) == false {
		http.Error(w, "blocked endpoint", 404)
		return
	}
	log.Println(fmt.Sprintf("tunnel to %s", r.Host))

	proxy := calculateProxy(r)
	raddr, err := net.ResolveTCPAddr("tcp", proxy)
	if err != nil {
		http.Error(w, "Unable to connect backconnect proxy", 404)
		return
	}
	destConn, err := net.DialTCP("tcp", nil, raddr)
	_, err = writeTCPConnection(destConn, r)

	// basically reading the 200 response and throwing it away
	_, err = readTCPResponse(destConn)

	// log out the request details
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go transfer(clientConn, destConn)
	go transfer(destConn, clientConn)
}

// copy data from source to destination
// Kudos to helinwang on gist https://gist.github.com/vmihailenco/1380352
func transfer(source io.ReadCloser, destination io.WriteCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

// get the proxy hint from the headers and use this to work out the proxy
// to use
func calculateProxy(req *http.Request) string {
	var proxy string
	hint := req.Header.Get("X-Proxy-Hint")
	switch hint {
	case "empty":
		break
	default:
		proxy = "175.45.134.96:80"
	}

	return proxy
}

// create a proxy url for use in the http transport
func proxyURL(req *http.Request) (*url.URL, error) {
	proxy := calculateProxy(req)
	return url.Parse(fmt.Sprintf("http://%s", proxy))
}

// create the http roundtripper transport for use in basic http requests
func createTransport() *http.Transport {
	transport := &http.Transport{
		Proxy: proxyURL,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
	}

	return transport
}

// handle basic HTTP traffic
func handleHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}

	transport := createTransport()
	resp, err := transport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// copy the header from source to destination
func copyHeader(src, dst http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// default route handler
func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			handleTunneling(w, r)
		} else {
			handleHTTP(w, r)
		}
	})
}

// ListenAndServeProxy will start the proxy on the local machine
func ListenAndServeProxy() {
	server := http.Server{
		Addr:    ":9001",
		Handler: handler(),
		// Disable HTTP/2.
		//TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	err := server.ListenAndServe()
	log.Fatal(err)
	panic(err)
}
