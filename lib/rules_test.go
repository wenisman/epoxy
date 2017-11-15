package lib

import (
	"log"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func setupViper() {
	viper.Set("proxies", map[string]interface{}{
		"192.168.100.1:80": 1,
		"192.168.100.2:80": 2,
	})
}

// TestFilterPriorityProxies will assert the values of the priority filter
func TestFilterPriorityProxies(t *testing.T) {
	start := time.Now()
	defer log.Println("filterPriorityProxy rule time taken:", time.Since(start))

	ph := &ProxyHint{}

	proxies := map[string]interface{}{
		"192.168.100.1:80": 1,
		"192.168.100.2:80": 2,
	}

	output := filterPriorityProxies(ph, proxies)

	if len(output) == 0 {
		t.Fatalf("No proxies returned")
	}

	if len(output) != 1 {
		t.Fatalf("default proxy priority not honoured")
	}

	ph.Priority = 2
	output = filterPriorityProxies(ph, proxies)

	if len(output) != 1 {
		t.Fatalf("Incorrect number of priority proxies returned")
	}

	for k := range output {
		if k != "192.168.100.2:80" {
			t.Fatalf("incorrect proxy returned for priority 2")
		}
	}
}

func TestFilterUseProxy(t *testing.T) {
	start := time.Now()
	defer log.Println("filterUseProxy rule time taken:", time.Since(start))

	proxies := map[string]interface{}{
		"192.168.100.1:80": 1,
		"192.168.100.2:80": 2,
	}

	ph := &ProxyHint{
		Use: "192.168.100.1:80",
	}

	proxy := filterUseProxy(ph, proxies)
	for k := range proxy {
		if k != "192.168.100.1:80" {
			t.Fatalf("incorrect proxy returned for use proxy")
		}
	}
}

func TestFailedProxyFilter(t *testing.T) {
	start := time.Now()
	defer log.Println("filterFailedProxy rule time taken:", time.Since(start))

	proxies := map[string]interface{}{
		"192.168.100.1:80": 1,
		"192.168.100.2:80": 2,
	}

	ph := &ProxyHint{
		Failed: []string{"192.168.100.1:80"},
	}

	output := filterFailedProxies(ph, proxies)

	if len(output) != 1 {
		t.Fatalf("Incorrect number of priority proxies returned")
	}

	for k := range output {
		if k != "192.168.100.2:80" {
			t.Fatalf("incorrect proxy returned for priority 2")
		}
	}

}
