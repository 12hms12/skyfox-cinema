package main

import (
	"flag"
	"fmt"
	"os"
	"skyfox/app/server"
	"skyfox/config"
)

const (
	configFileKey     = "configFile"
	defaultConfigFile = ""
	configFileUsage   = "/path/to/configfile/wrto/pwd"
)

// @title Booking API
// @version 1.0
// @description This is a skyfox

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /
//
//	@securityDefinitions.basic	BasicAuth
func main() {
	var configFile string

	flag.StringVar(&configFile, configFileKey, defaultConfigFile, configFileUsage)
	flag.Parse()

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = server.Init(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
