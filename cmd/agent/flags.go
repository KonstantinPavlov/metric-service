package main

import "flag"

var flagServerAddr string
var flagReportInterval int
var flagPollInterval int

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port metric server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()
}
