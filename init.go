package main

import (
	"flag"
)

func init() {
	flag.StringVar(&BindAddress, "server", "0.0.0.0:4050", "bot server port")
	flag.StringVar(&DatabaseType, "datastore", "sqlite3", "database type")
	flag.StringVar(&DatabaseURL, "datastoreurl", "go-neb.db?_busy_timeout=5000", "database url")
	flag.StringVar(&BaseURL, "baseurl", "http://localhost:4050", "base url")

	ConfigFile = "./data/config.yaml"

	flag.Parse()
}
