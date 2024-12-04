package main

import (
	"fmt"
	"log"
	"net"

	"pr10/pkg/services"
	"pr10/pkg/session"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalln("cant listen port", err)
	}

	server := grpc.NewServer()

	session.RegisterAuthCheckerServer(server, services.NewSessionManager())

	fmt.Println("starting server at :8081")
	server.Serve(lis)
}
