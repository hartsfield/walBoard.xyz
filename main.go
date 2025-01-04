package main // viewData represents the root model used to dynamically update the app

import (
	"fmt"
	"html/template"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	template.Must(templates.ParseGlob("internal/*/*/*"))
}

func main() {
	f := setupLogging()
	defer f.Close()

	ctx, srv := bolt()

	beginCache()

	log.Println("Waiting for connections @ http://localhost" + srv.Addr)
	fmt.Println("\n\nWaiting for connections @ http://localhost" + srv.Addr + "  -->  " + appConf.App.DomainName)

	<-ctx.Done()
}
