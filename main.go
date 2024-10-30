package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ohohestudio/sogorro/libs"
	"ohohestudio/sogorro/metadata"
	"os"
	"os/signal"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

type App struct {
	*http.Server
	ctx         context.Context
	fs          *firestore.Client
	projectId   string
	makeRequest func(method, url string, headers map[string]string, payload interface{}) ([]byte, error)
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

	app.makeRequest = libs.MakeRequest

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

func (a *App) findStation(w http.ResponseWriter, r *http.Request) {
	var webhookPayload struct {
		Destination string              `json:"destination"`
		Events      []libs.WebhookEvent `json:"events"`
	}

	err := json.NewDecoder(r.Body).Decode(&webhookPayload)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode JSON string: %v", err), http.StatusInternalServerError)
		return
	}

	payload := struct {
		To       string        `json:"to"`
		Messages []interface{} `json:"messages"`
	}{}

	payload.To = webhookPayload.Events[0].Source.UserId
	if webhookPayload.Events[0].Message.Type == "location" {
		latitude := webhookPayload.Events[0].Message.Latitude
		longitude := webhookPayload.Events[0].Message.Longitude
		query := a.fs.Collection("stations").Where("latitude", ">=", latitude-0.035).
			Where("latitude", "<=", latitude+0.035).
			Where("longitude", ">=", longitude-0.035).
			Where("longitude", "<=", longitude+0.035).
			Where("state", "==", 1)
		iter := query.Documents(a.ctx)

		var stations []libs.GoStation
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				fmt.Printf("failed to iterate document: %v\n", err)
			}

			stations = append(stations, libs.GoStation{
				Address:   doc.Data()["address"].(string),
				City:      doc.Data()["city"].(string),
				Distance:  libs.Haversine(doc.Data()["latitude"].(float64), doc.Data()["longitude"].(float64), latitude, longitude),
				District:  doc.Data()["district"].(string),
				Location:  doc.Data()["location"].(string),
				Latitude:  doc.Data()["latitude"].(float64),
				Longitude: doc.Data()["longitude"].(float64),
				VMType:    doc.Data()["vmType"].(int64),
			})
		}

		if len(stations) > 0 {
			sort.Slice(stations, func(j, k int) bool {
				return stations[j].Distance < stations[k].Distance
			})

			for _, station := range stations[:3] {
				payload.Messages = append(payload.Messages, libs.BubbleMessage(station))
			}
		} else {
			payload.Messages = append(payload.Messages, map[string]string{
				"type": "text",
				"text": "抱歉，您附近沒有找到 Gogoro 充電站。請嘗試分享其他位置或稍後再試。",
			})
		}
	} else {
		payload.Messages = append(payload.Messages, map[string]interface{}{
			"type":       "text",
			"text":       "歡迎使用 sogorro \n\n只要分享您的目前位置，我們會為您找到離您最近的 GoStation，方便您快速找到充電站！隨時隨地，讓騎乘更輕鬆愜意！",
			"quickReply": libs.WelcomeQuickReplyMessage(),
		})
	}

	result, err := a.makeRequest(
		http.MethodPost,
		os.Getenv("LINE_API_ENDPOINT"),
		map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", os.Getenv("LINEBOT_ACCESS_TOKEN")),
		},
		payload,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to push line message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
