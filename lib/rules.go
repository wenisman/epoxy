package lib

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/spf13/viper"
)

func isAllowedEndpoint(endpoint string) bool {
	start := time.Now()
	defer log.Println("endpoint rule time taken:", time.Since(start))
	blacklist := viper.GetStringSlice("blacklist")
	for _, v := range blacklist {
		result, _ := regexp.Match(v, []byte(endpoint))
		if result == true {
			return false
		}
	}
	return true
}

// ProxyHint structure that will define the structure of the header returned
// to us for hinting
type ProxyHint struct {
	Use      string   `json:"use" structs:"use" mapstructure:"use"`
	Failed   []string `json:"failed" structs:"failed" mapstructure:"failed"`
	Priority int      `json:"priority" structs:"priority" mapstructure:"priority"`
}

func filterUseProxy(ph *ProxyHint) map[string]interface{} {
	proxies := viper.GetStringMap("proxies")

	for _, proxy := range proxies {
		v := proxy.(map[string]interface{})
		if v["uri"] == ph.Use {
			return proxy.(map[string]interface{})
		}
	}

	return nil
}

func filterFailedProxies(ph *ProxyHint) map[string]interface{} {
	proxies := viper.GetStringMap("proxies")

	for _, p := range ph.Failed {
		for k, i := range proxies {
			v := i.(map[string]interface{})
			if v["uri"] == p {
				delete(proxies, k)
			}
		}
	}

	return proxies
}

func filterPriorityProxies(ph *ProxyHint, proxies map[string]interface{}) map[string]interface{} {
	if ph.Priority != 0 {
		priority := ph.Priority

		for k, v := range proxies {
			p := v.(map[string]interface{})
			if p["priority"] != priority {
				delete(proxies, k)
			}
		}
	}

	return proxies
}

func getProxy(req *http.Request) string {
	start := time.Now()
	defer log.Println("getProxy rule time taken:", time.Since(start))

	var hint ProxyHint
	hintHeader := req.Header.Get("X-Proxy-Hint")
	json.Unmarshal([]byte(hintHeader), &hint)

	proxy := filterUseProxy(&hint)
	if proxy != nil {
		return proxy["uri"].(string)
	}

	proxies := filterFailedProxies(&hint)
	proxies = filterPriorityProxies(&hint, proxies)

	var uris []string
	for _, v := range proxies {
		uris[len(uris)] = v.(map[string]interface{})["uri"].(string)
	}

	return uris[0]
}
