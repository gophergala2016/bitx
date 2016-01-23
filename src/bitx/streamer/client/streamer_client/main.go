package main

import (
	"flag"
	"log"

	"bitx/streamer/client"
)

var address = flag.String("address", "", "Address of streamer server")
var pair = flag.String("pair", "", "Market to stream, e.g. XBTZAR")

func main() {
	flag.Parse()

	cl := client.New(*pair)

	err := cl.Connect(*address)
	if err != nil {
		log.Fatal(err)
	}

	go cl.Stream()

	err = cl.FetchOrderBook()
	if err != nil {
		log.Fatal(err)
	}

	<-cl.Quit
	log.Printf("streamer_client: Exiting.")
}
