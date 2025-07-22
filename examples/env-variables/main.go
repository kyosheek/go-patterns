package main

import (
	"log"

	"kyoshee/singleton"
)

type config struct {
	dbURL  string
	apiURL string
}

func newConfig() *config {
	return &config{
		dbURL:  "postgresql://username:password@host:port/database",
		apiURL: "http://other-server:other-port/v1",
	}
}

var cfg *singleton.Singleton[config]

func init() {
	cfg = singleton.New(newConfig)
}

func main() {
	log.Printf("connecting to database '%s'", cfg.Get().dbURL)
	log.Printf("fetching data from '%s'", cfg.Get().apiURL)
}
