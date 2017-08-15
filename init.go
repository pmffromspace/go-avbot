package main

import (
	"flag"
)

func init() {
	flag.StringVar(&BIND_ADDRESS, "server", "127.0.0.1:4050", "bot server port")
	flag.StringVar(&DATABASE_TYPE, "datastore", "sqlite3", "database type")
	flag.StringVar(&DATABASE_URL, "datastoreurl", "go-neb.db?_busy_timeout=5000", "database url")
	flag.StringVar(&BASE_URL, "baseurl", "http://localhost:4050", "base url")

	LOG_DIR = "./log"
	CONFIG_FILE = "./config.yaml"

	flag.Parse()
}
