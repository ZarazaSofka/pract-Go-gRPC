package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"pr10/pkg/handlers"
	"pr10/pkg/session"
	"pr10/pkg/time"
	t "time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var pathsMap = make(map[string]int)

func main() {
	zapCfg := zap.NewProductionConfig()
	zapCfg.OutputPaths = []string{
		"/var/log/pr10/pr10.log",
		"stdout",
	}
	zapLogger, err := zapCfg.Build()
	if err != nil {
		panic("Logger creating error")
	}
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()


	service1, err := grpc.Dial(
		"service1:8081",
		grpc.WithInsecure(),
	)
	if err != nil {
		logger.Fatal("cant connect to grpc")
	}
	defer service1.Close()
	logger.Info("Connected to service 1")

	service2, err := grpc.Dial(
		"service2:8082",
		grpc.WithInsecure(),
	)
	if err != nil {
		logger.Fatal("cant connect to grpc")
	}
	defer service2.Close()
	logger.Info("Connected to service 2")

	sh := handlers.SessionHandler{
		SessManager: session.NewAuthCheckerClient(service1),
		Logger: logger,
	}

	th := handlers.TimeHandler{
		TimeService: time.NewTimeServiceClient(service2),
		Logger: logger,
	}
	r := http.NewServeMux()
	r.HandleFunc("/", sh.InnerPage)
	r.HandleFunc("/login", sh.LoginPage)
	r.HandleFunc("/logout", sh.LogoutPage)
	r.HandleFunc("/time", th.GetTime)
	http.Handle("/", updatePathsMap(r))

	go func() {
		for {
			now := t.Now()
			nextMidnight := t.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			durationUntilMidnight := nextMidnight.Sub(now)

			t.Sleep(durationUntilMidnight)
			sendAnalytics(logger)
			go func() {
				ticker := t.NewTicker(24 * t.Hour)
				for range ticker.C {
					sendAnalytics(logger)
				}
			}()
		}
	}()


	logger.Info("Starting server at :8080")
	http.ListenAndServe(":8080", nil)
}


func sendAnalytics(logger *zap.SugaredLogger) {
	jsonData, err := json.Marshal(pathsMap)
	if err != nil {
		logger.Errorf("Error encoding json data for analytics: %v", pathsMap)
		return
	}
	req, err := http.NewRequest("POST", "http://analytics:8000/api/analytics", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Error creating request to analytics: %w", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Error making request to analytics: %w", err)
		return
	}
	defer resp.Body.Close()

	clear(pathsMap)
}

func updatePathsMap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		pathsMap[path]++
		next.ServeHTTP(w, r)
	})
}