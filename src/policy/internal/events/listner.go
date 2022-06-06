package event

import (
	"context"
	"emcopolicy/pkg/grpc"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func listenOne(addr string) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect %v", err)
	}
	client := events.NewEventsClient(conn)
	defer conn.Close()
	ListenForEvents(client)
}

func ListenForEvents(c events.EventsClient) {
	client, err := c.EventUpdate(context.Background(), &events.EventInitiate{ServerId: 1003})
	if err != nil || client == nil {
		log.Println("Couldn't connect to client: ", err)
		return
	}
	for {
		fmt.Println(client)
		m, err := client.Recv()
		log.Println("New event received")
		if err != nil {
			log.Printf("Its an error %v\n", err)
			break
		}
		log.Println(m)
		//Put to channel
	}
}

func ListenServerInitiated(addrs []string) {
	for _, addr := range addrs {
		listenOne(addr)
	}
}
