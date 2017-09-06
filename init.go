package main

import (
	"flag"
)

func init() {
	flag.StringVar(&BindAddress, "server", "127.0.0.1:4050", "bot server port")
	flag.StringVar(&DatabaseType, "datastore", "sqlite3", "database type")
	flag.StringVar(&DatabaseUrl, "datastoreurl", "go-neb.db?_busy_timeout=5000", "database url")
	flag.StringVar(&BaseUrl, "baseurl", "http://localhost:4050", "base url")

	LogDir = "./log"
	ConfigFile = "./data/config.yaml"

	flag.Parse()
}
