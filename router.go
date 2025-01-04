package main

import "net/http"

// registerRoutes registers the routes with the provided *http.ServeMux
func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", home)
	mux.HandleFunc("/ranked", getByRanked)
	mux.HandleFunc("/chron", getByChron)
	mux.HandleFunc("/post/", viewPost)
	mux.HandleFunc("/submitForm", handleForm)
}
