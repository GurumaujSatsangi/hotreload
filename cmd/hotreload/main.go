package main

import (
	"flag"
)

type Config struct {
	Root  string
	Build string
	Exec  string
}

func main() {
	cfg := Config{}

	flag.StringVar(&cfg.Root, "root", ".", "")
	flag.StringVar(&cfg.Build, "build", "", "")
	flag.StringVar(&cfg.Exec, "exec", "", "")
	flag.Parse()
}
