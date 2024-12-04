package services

import (
	"context"
	pb "pr10/pkg/time"
	"time"
)

type TimeService struct {
	pb.UnimplementedTimeServiceServer
}

func (s TimeService) GetCurrentTime(ctx context.Context, request *pb.Empty) (*pb.TimeResponse, error) {
	currentTime := time.Now().Format(time.RFC3339)
	return &pb.TimeResponse{CurrentTime: currentTime}, nil
}