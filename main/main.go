package main

import (
	"net/http"

	"goblogengine/appenv"
	"goblogengine/blog"

	"goblogengine/external/github.com/gorilla/mux"
	"google.golang.org/appengine"
)

func main() {
	// Load the config
	err := appenv.Init()
	if err != nil {
		panic(err)
	}

	// Set the routes for the HTTP server
	r := mux.NewRouter()
	blog.Init(r)
	http.Handle("/", r)

	appengine.Main()
}
