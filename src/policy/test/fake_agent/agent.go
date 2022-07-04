package main

import (
	"emcopolicy/pkg/grpc"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	"net"
	"time"
)

const (
//addr = "127.0.0.1:9090"
)

type Server2 struct {
	events.EventsServer
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:9091")
	//lis, err := net.Listen("tcp", "172.31.82.234:9091")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}
	s := grpc.NewServer()
	events.RegisterEventsServer(s, &Server2{})
	err = s.Serve(lis)
	if err != nil {
		log.Println("Failed to serve")
		return
	}
}

func (Server2) EventUpdate(_ *events.ServerSpec, s events.Events_EventUpdateServer) error {
	var v = &events.AgentSpec{AgentId: "id4"}
	var m = &events.AgentMessage{
		Method: "PUT",
		Owner:  "bob@hooli.com",
		Path: []string{
			"pets",
			"pet113-987",
		},
		User: "bob@hooli.com",
	}
	spec, _ := anypb.New(v)
	msg, _ := anypb.New(m)
	for i := 40; i < 60; i++ {
		err := s.Send(&events.Event{
			EventId: "event1",
			AgentId: "id4",
			Spec:    spec,
			Message: msg,
		})

		//{EventId: strconv.Itoa(i)})
		ctx := s.Context()
		fmt.Println("Context = ", ctx)
		if err != nil {
			fmt.Println("received err", err)
			break
		}
		time.Sleep(10 * time.Second)
	}
	return nil
}
