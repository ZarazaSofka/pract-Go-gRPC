package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	pb "pr10/pkg/time"

	"go.uber.org/zap"
)

type TimeHandler struct {
	TimeService pb.TimeServiceClient
	Logger *zap.SugaredLogger
}

func (th *TimeHandler) GetTime(w http.ResponseWriter, r *http.Request) {
	response, err := th.TimeService.GetCurrentTime(context.Background(), &pb.Empty{})
	if err != nil {
		th.Logger.Error("Failed to get current time from gRPC server")
		http.Error(w, "Failed to get current time from gRPC server", http.StatusInternalServerError)
		return
	}

	timeJSON, err := json.Marshal(map[string]string{"Текущее время": response.CurrentTime})
	if err != nil {
		th.Logger.Error("Failed to marshal JSON response")
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	th.Logger.Info("Sending current time")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(timeJSON)
}