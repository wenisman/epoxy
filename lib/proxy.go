package lib

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

// use tcp tunneling as for CONNECT as this covers https and ws
func handleTunneling(w http.ResponseWriter, r *http.Request) {
	raddr, err := net.ResolveTCPAddr("tcp", "175.45.134.96:80")
	if err != nil {
		log.Println("error resolving address", err)
	}
	destConn, err := net.DialTCP("tcp", nil, raddr)
	_, err = destConn.Write([]byte(fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: curl/7.56.1\r\nProxy-Connection: Keep-Alive\r\n\r\n", r.Host, r.Host)))

	// basically reading the 200 response and throwing it away
	tmp := make([]byte, 256)
	_, err = destConn.Read(tmp)
	if err != nil {
		log.Println(err)
	}

	// log out the request details
	if err != nil {
		log.Println("error proxying through proxy", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Println("error whilst hijacking", err)
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

func setProxy(req *http.Request) (*url.URL, error) {
	return url.Parse("http://139.59.2.223:8888")
}

func createTransport() *http.Transport {
	transport := &http.Transport{
		Proxy: setProxy,
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
	log.Println("handling http request", req.Host, req.URL.String())
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}

	transport := createTransport()
	resp, err := transport.RoundTrip(req)
	if err != nil {
		log.Println("error during http request", err)
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

	log.Fatal(server.ListenAndServe())
}
