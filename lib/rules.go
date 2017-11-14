package lib

import (
	"log"
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
