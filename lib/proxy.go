package lib

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

	proxy := getProxy(r)

	log.WithFields(log.Fields{
		"connection": "tunnel",
		"proxy":      proxy,
		"endpoint":   "r.Host",
	}).Debugf("tunnel to %s through", r.Host)

	if proxy == "" {
		// fall back to direct connection if no proxy is found
		proxy = r.Host
	}

	raddr, err := net.ResolveTCPAddr("tcp", proxy)
	if err != nil {
		http.Error(w, "Unable to connect backconnect proxy", 404)
		return
	}
	destConn, err := net.DialTCP("tcp", nil, raddr)
	_, err = writeTCPConnection(destConn, r)

	// reading the response from remote
	resp, err := readTCPResponse(destConn)

	// log out the request details
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// pass the response back to the client so they can decide what to do
	w.Write(resp)
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

// create a proxy url for use in the http transport
func proxyURL(req *http.Request) (*url.URL, error) {
	proxy := getProxy(req)
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

	log.WithFields(log.Fields{
		"endpoint":   "r.Host",
		"connection": "http",
	}).Debugf("http connection to %s", req.Host)

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
	port := viper.GetInt("port")

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler(),
		// Disable HTTP/2.
		//TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	err := server.ListenAndServe()
	log.Fatal(err)
	panic(err)
}
