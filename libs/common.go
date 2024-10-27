package libs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

type GoStation struct {
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Distance  float64 `json:"distance"`
	District  string  `json:"district"`
	Location  string  `json:"location"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	VMType    int64   `json:"vmType"`
}

// 計算距離
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Radius of Earth in kilometers
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	ax := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(ax), math.Sqrt(1-ax))

	distance := R * c
	return distance
}

func MakeRequest(method, url string, headers map[string]string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to encoding object: %v", err)
	}
	fmt.Println(string(jsonData))

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

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
