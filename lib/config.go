package lib

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// SetLogLevel will set the logging level for your application
func SetLogLevel() {
	if viper.InConfig("loglevel") && viper.GetString("loglevel") != "" {
		level, err := log.ParseLevel(viper.GetString("loglevel"))
		if err != nil {
			panic(err)
		}
		log.SetLevel(level)
	}
}

func unmarshalConfig(name string, object interface{}) {
	if reflect.TypeOf(viper.Get(name)).Kind() != reflect.TypeOf(object).Kind() {
		yaml.Unmarshal([]byte(viper.GetString(name)), &object)
		viper.Set(name, object)
	}
}

// LoadConfig will read the default and environment specific config files
func LoadConfig() {
	fmt.Println("loading config")
	viper.SetConfigType("yaml")

	// load the default config file
	viper.SetConfigFile("./config/default.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error loading config: %s", err))
	}

	// load the environment specific file
	if viper.GetString("environment") != "" {
		viper.SetConfigFile(fmt.Sprintf("./config/%s.yml", viper.GetString("environment")))
		viper.MergeInConfig()
	}

	// load in the proxies from the environment or comamnd line
	unmarshalConfig("proxies", make(map[string]interface{}))

	// check we are a slice, else convert the blacklist data
	unmarshalConfig("blacklist", []string{})
}
