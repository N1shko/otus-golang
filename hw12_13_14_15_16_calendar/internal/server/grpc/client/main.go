package main

import (
	"context"
	"log"
	"time"

	pb "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.NewClient("localhost:10001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create the EventService client
	client := pb.NewEventServiceClient(conn)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Call the ListEvents RPC
	resp, err := client.ListEvents(ctx, &pb.ListEventsRequest{})
	if err != nil {
		log.Printf("Failed to call ListEvents: %v", err)
		return
	}

	// Print the response
	for i, event := range resp.GetEvents() {
		log.Printf("Event %d:", i+1)
		log.Printf("  ID: %s", event.GetId())
		log.Printf("  Title: %s", event.GetTitle())
		log.Printf("  DateStart: %s", event.GetDateStart().AsTime().Format(time.RFC3339))
		log.Printf("  DateEnd: %s", event.GetDateEnd().AsTime().Format(time.RFC3339))
		log.Printf("  Description: %s", event.GetDescription())
		log.Printf("  UserID: %s", event.GetUserId())
		log.Printf("  SendBefore: %s", event.GetSendBefore().AsTime().Format(time.RFC3339))
	}
}
