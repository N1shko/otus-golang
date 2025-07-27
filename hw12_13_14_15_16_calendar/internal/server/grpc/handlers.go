package internalgrpc

import (
	"context"

	pb "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	message := "Hello, " + req.Name + "!"
	return &pb.HelloResponse{Message: message}, nil
}

// func (s *Server) AddEvent(ctx context.Context, req *pb.AddEventRequest) (*pb.AddEventResponse, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// }

// func (s *Server) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// }

// func (s *Server) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// }

func (s *Server) ListEvents(ctx context.Context, _ *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	events, err := s.app.ListEvents(ctx)
	if err != nil {
		s.logger.Error("Failed to list events", "error", err.Error())
		return nil, err
	}
	pbEvents := make([]*pb.Event, 0, len(events))
	for _, event := range events {
		pbEvents = append(pbEvents, toProtoEvent(event))
	}
	return &pb.ListEventsResponse{
		Events: pbEvents,
	}, nil
}

// func (s *Server) GetEventsByTimeRange(ctx context.Context, req *pb.GetEventsByTimeRangeRequest)
// (*pb.GetEventsByTimeRangeResponse, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// }

func toProtoEvent(e storage.Event) *pb.Event {
	return &pb.Event{
		Id:          e.ID.String(),
		Title:       e.Title,
		DateStart:   timestamppb.New(e.DateStart),
		DateEnd:     timestamppb.New(e.DateEnd),
		Description: e.Description,
		UserId:      e.UserID,
		SendBefore:  timestamppb.New(e.SendBefore),
	}
}

func fromProtoEvent(pe *pb.Event) (storage.Event, error) {
	id, err := uuid.Parse(pe.Id)
	if err != nil {
		return storage.Event{}, err
	}

	return storage.Event{
		ID:          id,
		Title:       pe.Title,
		DateStart:   pe.DateStart.AsTime(),
		DateEnd:     pe.DateEnd.AsTime(),
		Description: pe.Description,
		UserID:      pe.UserId,
		SendBefore:  pe.SendBefore.AsTime(),
	}, nil
}
