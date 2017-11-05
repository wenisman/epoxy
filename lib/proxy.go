package lib

import (
	"time"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
)

// use tcp tunneling as for CONNECT as this covers https and wss
func handleTunneling(w http.ResponseWriter, r *http.Request) {
	//dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	
	// change to use a new destination proxy 
	// TODO : validate the request host is a fully qualified address
	host := "https://"+ r.URL.Host
	log.Println("request host", host)
	// target := "213.136.105.62:443"
	// raddr,_ := net.ResolveTCPAddr("tcp",target)
	// dest_conn, err := net.DialTCP("tcp", nil, raddr )
	dest_conn, err := net.DialTimeout("tcp", "213.136.105.62:443", 10*time.Second)
	defer dest_conn.Close();	

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
	client_conn, _, err := hijacker.Hijack()
	defer client_conn.Close()
	if err != nil {
		log.Println("error whilst hijacking", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(client_conn, dest_conn)
	go transfer(dest_conn, client_conn)
}

// copy data from source to destination
func transfer(source io.ReadCloser, destination io.WriteCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

// handle basic HTTP traffic
func handleHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("handling http request")	
	resp, err := http.DefaultTransport.RoundTrip(req)
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

func ListenAndServeProxy() {
	server := http.Server{
		Addr: ":9001",
		Handler: handler(),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Fatal(server.ListenAndServe())
}