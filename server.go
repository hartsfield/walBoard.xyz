package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func bolt() (ctx context.Context, srv *http.Server) {
	var mux *http.ServeMux = http.NewServeMux()
	registerRoutes(mux)

	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	srv = serverFromConf(mux)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Panicln(err)
		}
		cancelCtx()
	}()

	return
}

// serverFromConf returns a *http.Server with a pre-defined configuration
func serverFromConf(mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:              servicePort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       5 * time.Second,
	}
}
