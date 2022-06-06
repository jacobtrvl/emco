package main

import (
	"emcopolicy/pkg/grpc"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
	"time"
)

const (
//addr = "127.0.0.1:9090"
)

type Server struct {
	events.EventsServer
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:9090")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}
	s := grpc.NewServer()
	events.RegisterEventsServer(s, &Server{})
	err = s.Serve(lis)
	if err != nil {
		log.Println("Failed to serve")
		return
	}
}

func (Server) EventUpdate(_ *events.EventInitiate, s events.Events_EventUpdateServer) error {
	for i := 0; i < 20; i++ {
		err := s.Send(&events.Event{EventId: strconv.Itoa(i)})
		ctx := s.Context()
		fmt.Println("Context = ", ctx)
		if err != nil {
			fmt.Println("received err", err)
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
