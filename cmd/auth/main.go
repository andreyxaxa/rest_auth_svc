package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/andreyxaxa/rest_auth_svc/internal/app/server"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "configs/server.toml", "path to config file")
}

func main() {
	flag.Parse()

	config := server.NewConfig()
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Start(config); err != nil {
		log.Fatal(err)
	}
}
