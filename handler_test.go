package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
)

// MockFirestoreClient using testify's mock package
type MockFirestoreClient struct {
	mock.Mock
}

// Mock Collection method
func (m *MockFirestoreClient) Collection(path string) *firestore.CollectionRef {
	args := m.Called(path)
	return args.Get(0).(*firestore.CollectionRef)
}

// Mock DocumentIterator for Documents method
type MockDocumentIterator struct {
	mock.Mock
	Docs []firestore.DocumentSnapshot
	Pos  int
}

// Mock Firestore document snapshot
type MockDocumentSnapshot struct {
	mock.Mock
	DataMap map[string]interface{}
}

func (m *MockDocumentSnapshot) Data() map[string]interface{} {
	return m.DataMap
}

// Mock the Documents method
func (m *MockFirestoreClient) Documents(ctx context.Context) *MockDocumentIterator {
	args := m.Called(ctx)
	return args.Get(0).(*MockDocumentIterator)
}

// Mock Next method for DocumentIterator
func (m *MockDocumentIterator) Next() (*firestore.DocumentSnapshot, error) {
	args := m.Called()
	if m.Pos >= len(m.Docs) {
		return nil, iterator.Done
	}
	doc := &m.Docs[m.Pos]
	m.Pos++
	return doc, args.Error(1)
}

func TestFindStation(t *testing.T) {
	mockFirestore := new(MockFirestoreClient)
	mockDocIter := new(MockDocumentIterator)
	mockDoc := &MockDocumentSnapshot{
		DataMap: map[string]interface{}{
			"latitude":  12.34,
			"longitude": 56.78,
			"city":      "CityName",
			"district":  "DistrictName",
			"rId":       "station1",
		},
	}

	// Setup mock expectations
	mockDocIter.On("Next").Return(mockDoc, nil).Once()
	mockDocIter.On("Next").Return(nil, iterator.Done).Once()
	mockFirestore.On("Documents", mock.Anything).Return(mockDocIter)

	app := &App{
		fs:  mockFirestore,
		ctx: context.Background(),
	}

	linePayload := map[string]interface{}{
		"destination": "My Destination",
		"events": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"latitude":  12.34,
					"longitude": 56.78,
				},
			},
		},
	}

	// Convert the linePayload map to JSON
	body, _ := json.Marshal(linePayload)

	// Create a mock HTTP request
	req := httptest.NewRequest("POST", "/findStation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the findStation handler function
	app.findStation(rr, req)

	// Check the status code is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the Content-Type header
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Decode the response body to validate the returned data
	var stations map[string]map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&stations)
	assert.NoError(t, err)
}
