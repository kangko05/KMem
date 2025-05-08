package main

import (
	"context"
	"file-service/internal/server"
	"log"
)

const PORT = ":8001"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("running file-service at %s\n", PORT)

	serv, err := server.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	serv.Run()
}
