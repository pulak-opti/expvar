package main

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	// Create a new Chi router
	r := chi.NewRouter()

	counter := expvar.NewInt("counter")
	counter.Add(1)
	sample := expvar.NewString("name")
	sample.Set("value")

	// Register middleware
	r.Use(middleware.Logger)

	// Register expvar handler
	r.Get("/debug/vars", expvarHandler)

	// Add your other routes and handlers
	r.Get("/", helloHandler)

	// Start the server
	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", r)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	expvar.Handler().ServeHTTP(w, r)
}
