package main

import (
	"flag"
	"log"
	"net/http"
    "fmt"
)

func main() {

	// command line flags
	host := flag.String	("h", "192.168.1.110", "host server")
	port := flag.Int	("p", 80, "port to serve on")
	dir  := flag.String	("dir", "./", "directory of web files")
	flag.Parse()

	// handle all requests by serving a file of the same name
	fs 	:= http.Dir(*dir)
	fileHandler := http.FileServer(fs)
	http.Handle("/", fileHandler)

	log.Printf("Running on port %d\n", *port)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	// this call blocks -- the progam runs here forever
	err := http.ListenAndServe(addr, nil)
	fmt.Println(err.Error())

}
