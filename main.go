package main

import (
	"expvar"
	"fmt"
	"net/http"
)

func main() {
	// Define some variables to export
	counter := expvar.NewInt("counter")
	counter.Add(1)

	// Start an HTTP server
	http.HandleFunc("/hello", helloHandler)
	//http.Handle("/debug/vars", expvar.Handler())
	fmt.Println("server is running on 8080")
	http.ListenAndServe(":8080", nil)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}
