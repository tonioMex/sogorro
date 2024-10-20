package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ohohestudio/sogorro/metadata"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	*http.Server
	ctx       context.Context
	fs        *firestore.Client
	logger    *logging.Logger
	projectId string
}

type LineEvent struct {
	Type    string `json:"type"`
	Message struct {
		Type      string  `json:"type"`
		Id        string  `json:"id"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Address   string  `json:"address"`
	} `json:"message"`
	WebhookEventId  string `json:"webhookEventId"`
	DeliveryContext struct {
		IsRedelivery bool `json:"isRedelivery"`
	}
	Timestamp int64 `json:"timestamp"`
	Source    struct {
		Type   string `json:"type"`
		UserId string `json:"userId"`
	}
	ReplyToken string `json:"replyToken"`
	Mode       string `json:"active"`
}

func main() {
	ctx := context.Background()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	projectId := os.Getenv("GOOGLE_CLOUD_PROJECT")

	app, err := newApp(ctx, port, projectId)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("start sogorro API server, running on port: %s\n", port)

	go func() {
		if err := app.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("sogorro API server closed unexpectedly: %v", err)
		}
	}()

	nofityCtx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	<-nofityCtx.Done()
	log.Println("manually shutdown sogorro API server")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	app.Shutdown(ctx)
	log.Println("sogorro API server has been shutdown")
}

func newApp(ctx context.Context, port, projectId string) (*App, error) {
	app := &App{
		ctx: ctx,
		Server: &http.Server{
			Addr:           fmt.Sprintf(":%s", port),
			ReadTimeout:    15 * time.Second,
			WriteTimeout:   15 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	// Get Project ID
	if projectId == "" {
		projId, err := metadata.ProjectId(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to detect Project ID from GOOGLE_CLOUD_PROJECT or Google metadata server: %v", err)
		}
		projectId = projId
	}
	app.projectId = projectId

	// firestore
	fsClient, err := getFirebaseClient(ctx, app.projectId)
	if err != nil {
		return nil, err
	}
	app.fs = fsClient

	// Logging
	logger, err := getLogger(ctx, app.projectId)
	if err != nil {
		return nil, err
	}
	app.logger = logger

	// Router
	r := mux.NewRouter()
	r.HandleFunc("/station", app.findStation).Methods("POST")
	app.Handler = r

	return app, nil
}

func getFirebaseClient(ctx context.Context, projectId string) (*firestore.Client, error) {
	config := &firebase.Config{
		ProjectID: projectId,
	}

	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firebase app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firestore service: %v", err)
	}

	return client, nil
}

func getLogger(ctx context.Context, projectId string) (*logging.Logger, error) {
	client, err := logging.NewClient(
		ctx,
		fmt.Sprintf("projects/%s", projectId),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize logging client: %v", err)
	}

	return client.Logger(fmt.Sprintf("%s-logger", projectId), logging.RedirectAsJSON(os.Stderr)), nil
}

func (a *App) makeRequest(method, url string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}

	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	return body, nil
}

func (a *App) findStation(w http.ResponseWriter, r *http.Request) {
	var linePayload struct {
		Destination string      `json:"destination"`
		Events      []LineEvent `json:"events"`
	}

	err := json.NewDecoder(r.Body).Decode(&linePayload)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode JSON string: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	m := make(map[string]float64)
	m["latitude"] = linePayload.Events[0].Message.Latitude
	m["longitude"] = linePayload.Events[0].Message.Longitude

	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, "failed to encode object", http.StatusInternalServerError)
		return
	}
	// a.log.Log(logging.Entry{
	// 	Severity: logging.Info,
	// 	HTTPRequest: &logging.HTTPRequest{
	// 		Request: r,
	// 	},
	// 	Labels:  map[string]string{"arbitraryField": "custom entry"},
	// 	Payload: "Structured logging example",
	// })
}
