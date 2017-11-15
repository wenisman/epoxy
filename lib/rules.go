package lib

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// check that the end point is allowed to be reached based on the rules
func isAllowedEndpoint(endpoint string) bool {
	start := time.Now()
	defer log.WithFields(log.Fields{
		"rule": "isAllowedEndpoint",
	}).Debugf("time taken: %s", time.Since(start))

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

// this will validate the proxy specified in the hint is still valid
func filterUseProxy(ph *ProxyHint, proxies map[string]interface{}) map[string]interface{} {
	start := time.Now()
	defer log.WithFields(log.Fields{
		"rule": "filterUseProxy",
	}).Debugf("time taken: %s", time.Since(start))

	if proxies[ph.Use] != nil {
		return map[string]interface{}{
			ph.Use: proxies[ph.Use],
		}
	}

	return nil
}

// remove all the failed proxies from the list of available proxies
func filterFailedProxies(ph *ProxyHint, proxies map[string]interface{}) map[string]interface{} {
	start := time.Now()
	defer log.WithFields(log.Fields{
		"rule": "filterFailedProxies",
	}).Debugf("time taken: %s", time.Since(start))

	for _, p := range ph.Failed {
		delete(proxies, p)
	}

	return proxies
}

// remove all proxies from the list that are not of the desired priority
func filterPriorityProxies(ph *ProxyHint, proxies map[string]interface{}) map[string]interface{} {
	start := time.Now()
	defer log.WithFields(log.Fields{
		"rule": "filterPriorityProxies",
	}).Debugf("time taken: %s", time.Since(start))

	priority := 2
	if ph.Priority != 0 {
		priority = ph.Priority
	}

	// remove all the proxies with priority not matching the required
	for k, v := range proxies {
		if priority != v.(int) {
			delete(proxies, k)
		}
	}

	return proxies
}

func getProxy(req *http.Request) string {
	start := time.Now()
	defer log.WithFields(log.Fields{
		"rule": "getProxy",
	}).Debugf("time taken: %s", time.Since(start))

	proxies := viper.GetStringMap("proxies")

	var hint ProxyHint
	hintHeader := req.Header.Get("X-Proxy-Hint")
	if hintHeader != "" {
		// we have the header so process
		json.Unmarshal([]byte(hintHeader), &hint)

		if hint.Use != "" {
			return hint.Use
		}

		proxies = filterFailedProxies(&hint, proxies)
		proxies = filterPriorityProxies(&hint, proxies)
	}
	// extract the uris and return a random uri
	var uris []string
	for k := range proxies {
		uris = append(uris, k)
	}

	return uris[rand.Intn(len(uris))]
}
