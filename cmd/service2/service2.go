package main

import (
	"fmt"
	"log"
	"net"

	"pr10/pkg/services"
	"pr10/pkg/time"

	"google.golang.org/grpc"
)

func main() {
	ts := services.TimeService{}

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalln("cant listen port", err)
	}

	server := grpc.NewServer()

	time.RegisterTimeServiceServer(server, ts)

	fmt.Println("starting server at :8082")
	server.Serve(lis)
}
