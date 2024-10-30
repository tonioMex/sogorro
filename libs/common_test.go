package libs

import (
	"bytes"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockPayload struct {
	Message string `json:"message"`
}

func TestMakeRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		headers        map[string]string
		payload        interface{}
		mockServerFunc func(w http.ResponseWriter, r *http.Request)
		expectedError  string
		expectedBody   string
	}{
		{
			name:   "Successful GET request",
			method: http.MethodGet,
			url:    "/",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			payload: MockPayload{
				Message: "Hello world",
			},
			mockServerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "ok"}`))
			},
			expectedBody: `{"status": "ok"}`,
		},
		{
			name:   "HTTP client error",
			method: http.MethodGet,
			url:    "http://localhost:12345",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			payload: MockPayload{
				Message: "Hello world",
			},
			expectedError: "unable to make request",
		},
		{
			name:   "Read response body error",
			method: http.MethodGet,
			url:    "/",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			payload: MockPayload{
				Message: "Hello world",
			},
			mockServerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "1")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte{})
			},
			expectedError: "unable to read response body",
		},
		{
			name:          "encode payload error",
			method:        http.MethodPost,
			url:           "/",
			headers:       nil,
			payload:       make(chan int),
			expectedError: "unable to encode object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.mockServerFunc != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.mockServerFunc))
				defer server.Close()
				tt.url = server.URL + tt.url
			}

			body, err := MakeRequest(tt.method, tt.url, tt.headers, tt.payload)

			if err != nil {
				if tt.expectedError == "" || !bytes.Contains([]byte(err.Error()), []byte(tt.expectedError)) {
					t.Errorf("unexpected error: got %v, want %v", err, tt.expectedError)
				}
				return
			}

			if string(body) != tt.expectedBody {
				t.Errorf("unexpected body: got %s, but want %s", body, tt.expectedBody)
			}
		})
	}
}

func TestHaversine(t *testing.T) {
	tests := []struct {
		name             string
		lat1, lon1       float64
		lat2, lon2       float64
		expectedDistance float64
	}{
		{
			name:             "same location",
			lat1:             19.427050,
			lon1:             -99.127571,
			lat2:             19.427050,
			lon2:             -99.127571,
			expectedDistance: 0,
		},
		{
			name:             "Mexico City to Guadalajara",
			lat1:             19.427050,
			lon1:             -99.127571,
			lat2:             20.673590,
			lon2:             -103.343803,
			expectedDistance: 461.7,
		},
	}

	const tolerance = 0.1 // 精度 0.1 km
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calcDistance := Haversine(tt.lat1, tt.lon1, tt.lat2, tt.lon2)

			if math.Abs(calcDistance-tt.expectedDistance) > tolerance {
				t.Errorf("calculate by Haversine (%v), want %v (within tolerance %v)", calcDistance, tt.expectedDistance, tolerance)
			}
		})
	}
}
