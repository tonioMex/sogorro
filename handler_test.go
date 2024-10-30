package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"ohohestudio/sogorro/libs"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mockMakeRequest(method, url string, headers map[string]string, payload interface{}) ([]byte, error) {
	result := map[string]interface{}{
		"sentMessages": []map[string]string{
			{
				"id":         "message-id",
				"quoteToken": "message-quote-token",
			},
		},
	}

	return json.Marshal(result)
}

func mockWebhookEvent() libs.WebhookEvent {
	return libs.WebhookEvent{
		Type: "message",
		Message: struct {
			Type            string  "json:\"type\""
			Id              string  "json:\"id\""
			Latitude        float64 "json:\"latitude,omitempty\""
			Longitude       float64 "json:\"longitude,omitempty\""
			Address         string  "json:\"address,omitempty\""
			QuotedMessageId string  "json:\"quotedMessageId,omitempty\""
			QuoteToken      string  "json:\"quoteToken\""
			Text            string  "json:\"text,omitempty\""
		}{
			Type:       "text",
			Id:         "message-id",
			QuoteToken: "quote-token",
			Text:       "message text",
		},
		WebhookEventId: "webhook-event-id",
		DeliveryContext: struct {
			IsRedelivery bool "json:\"isRedelivery\""
		}{
			IsRedelivery: false,
		},
		Timestamp: time.Now().Unix(),
		Source: struct {
			Type   string "json:\"type\""
			UserId string "json:\"userId\""
		}{
			Type:   "user",
			UserId: "user-id",
		},
		ReplyToken: "reply-token",
		Mode:       "active",
	}
}

func TestFindStation(t *testing.T) {
	os.Setenv("LINE_API_ENDPOINT", "https://api.line.me/v2/bot/message/push")
	os.Setenv("LINEBOT_ACCESS_TOKEN", "mock-token")

	app := &App{
		ctx:         context.TODO(),
		makeRequest: mockMakeRequest,
	}

	tests := []struct {
		name           string
		webhookPayload interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "non-location message",
			webhookPayload: struct {
				Destination string
				Events      []libs.WebhookEvent
			}{
				Destination: "destination",
				Events:      []libs.WebhookEvent{mockWebhookEvent()},
			},
			expectedStatus: 200,
			expectedBody:   `{"sentMessages":[{"id":"message-id","quoteToken":"message-quote-token"}]}`,
		},
		{
			name:           "invalid line message",
			webhookPayload: `{"invalidKey":"invalidValue"}`,
			expectedStatus: 500,
			expectedBody:   "failed to decode JSON string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.webhookPayload)
			req := httptest.NewRequest(http.MethodPost, "/station", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			app.findStation(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}
