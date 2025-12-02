package main

import (
	"log"
	"net"

	"po5/internal/httpmini"
)

func main() {
	srv := &httpmini.Server{
		Addr:      ":7878",
		PublicDir: "public",
	}

	log.Printf("Listening on %s ...", srv.Addr)

	err := srv.ListenAndServe(func(conn net.Conn) {
		go srv.HandleConn(conn)
	})

	if err != nil {
		log.Fatal(err)
	}
}
