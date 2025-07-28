package internalgrpc

import (
	"context"
	"time"

	pb "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *Server) AddEvent(ctx context.Context, req *pb.AddEventRequest) (*pb.AddEventResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	event := storage.Event{
		Title:       req.Title,
		DateStart:   req.DateStart.AsTime(),
		DateEnd:     req.DateEnd.AsTime(),
		Description: req.Description,
		UserID:      req.UserId,
		SendBefore:  req.SendBefore.AsTime(),
	}

	if event.Title == "" || event.DateStart.IsZero() || event.DateEnd.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "Missing required fields")
	}

	if event.DateEnd.Before(event.DateStart) {
		return nil, status.Error(codes.InvalidArgument, "End date must be after start date")
	}

	event.ID = uuid.New()

	if err := s.app.AddEvent(ctx, event); err != nil {
		s.logger.Error("Failed to create event", "error", err.Error())
		return nil, status.Error(codes.Internal, "Failed to create event")
	}

	return &pb.AddEventResponse{
		Id: event.ID.String(),
	}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	idStr := req.GetId()
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	event := storage.Event{ID: id}

	if err := s.app.DeleteEvent(ctx, event); err != nil {
		s.logger.Error("Failed to delete event", "error", err.Error())
		return nil, err
	}
	return &pb.DeleteEventResponse{}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	id, err := uuid.Parse(req.Event.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid event ID")
	}
	existingEvent, err := s.app.GetEvent(ctx, id)
	if err != nil {
		s.logger.Error("Failed to fetch event", "error", err.Error())
		return nil, status.Error(codes.NotFound, "Event not found")
	}
	updatedEvent := existingEvent
	eventReq := req.Event
	if eventReq.Title != "" {
		updatedEvent.Title = eventReq.Title
	}
	if eventReq.DateStart != nil && !eventReq.DateStart.AsTime().IsZero() {
		updatedEvent.DateStart = eventReq.DateStart.AsTime()
	}
	if eventReq.DateEnd != nil && !eventReq.DateEnd.AsTime().IsZero() {
		updatedEvent.DateEnd = eventReq.DateEnd.AsTime()
	}
	if eventReq.Description != "" {
		updatedEvent.Description = eventReq.Description
	}
	if eventReq.UserId != "" {
		updatedEvent.UserID = eventReq.UserId
	}
	if eventReq.SendBefore != nil && !eventReq.SendBefore.AsTime().IsZero() {
		updatedEvent.SendBefore = eventReq.SendBefore.AsTime()
	}
	updatedEvent.ID = id
	if updatedEvent.Title == "" || updatedEvent.DateStart.IsZero() || updatedEvent.DateEnd.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "Missing required fields after update")
	}

	if updatedEvent.DateEnd.Before(updatedEvent.DateStart) {
		return nil, status.Error(codes.InvalidArgument, "End date must be after start date")
	}

	if err := s.app.UpdateEvent(ctx, updatedEvent); err != nil {
		s.logger.Error("Failed to update event", "error", err.Error())
		return nil, status.Error(codes.Internal, "Failed to update event")
	}
	return &pb.UpdateEventResponse{
		Id: id.String(),
	}, nil
}

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

func (s *Server) GetEventsByTimeRange(ctx context.Context,
	req *pb.GetEventsByTimeRangeRequest,
) (*pb.GetEventsByTimeRangeResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	rangeType := req.GetRangeType()
	startDate := req.GetStartDate()

	if startDate == nil {
		return nil, status.Error(codes.InvalidArgument, "Date cant be empty")
	}

	if !startDate.IsValid() || startDate.AsTime().IsZero() {
		return nil, status.Error(codes.InvalidArgument, "Invalid date format, use YYYY-MM-DD")
	}

	if rangeType == pb.TimeRangeType_TIME_RANGE_TYPE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "Range type must be specified")
	}

	startTime := startDate.AsTime()
	var endTime time.Time

	switch rangeType {
	case pb.TimeRangeType_DAY:
		endTime = startTime.Add(24 * time.Hour)
	case pb.TimeRangeType_WEEK:
		endTime = startTime.Add(7 * 24 * time.Hour)
	case pb.TimeRangeType_MONTH:
		endTime = startTime.AddDate(0, 1, 0)
	default:
		return nil, status.Error(codes.InvalidArgument, "Unsupported range type")
	}

	events, err := s.app.GetEventsByTimeRange(ctx, startTime, endTime)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to retrieve events")
	}

	pbEvents := make([]*pb.Event, 0, len(events))
	for _, event := range events {
		pbEvents = append(pbEvents, toProtoEvent(event))
	}
	return &pb.GetEventsByTimeRangeResponse{
		Events: pbEvents,
	}, nil
}

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

// func fromProtoEvent(pe *pb.Event) (storage.Event, error) {
// 	id, err := uuid.Parse(pe.Id)
// 	if err != nil {
// 		return storage.Event{}, err
// 	}

// 	return storage.Event{
// 		ID:          id,
// 		Title:       pe.Title,
// 		DateStart:   pe.DateStart.AsTime(),
// 		DateEnd:     pe.DateEnd.AsTime(),
// 		Description: pe.Description,
// 		UserID:      pe.UserId,
// 		SendBefore:  pe.SendBefore.AsTime(),
// 	}, nil
// }
