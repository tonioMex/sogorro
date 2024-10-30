package libs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBubbleMessage(t *testing.T) {
	tests := []struct {
		name    string
		station GoStation
	}{
		{
			name: "GoStation message",
			station: GoStation{
				Location:  "Station Ermita",
				Address:   "Calle 3 Pantitlan",
				VMType:    1,
				Distance:  3.16,
				Latitude:  19.427050,
				Longitude: -99.127571,
			},
		},
		{
			name: "SuperGoStation message",
			station: GoStation{
				Location:  "Station Pantitlan",
				Address:   "Rojo Gomes",
				VMType:    3,
				Distance:  1.28,
				Latitude:  20.673590,
				Longitude: -103.343803,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BubbleMessage(tt.station)
			fmt.Println((result))

			assert.Equal(t, "flex", result.Type)
			assert.Equal(t, "sogorro", result.AltText)
			assert.Equal(t, tt.station.Location, result.Contents.Body.Contents[0].(TextTemplate).Text)
			if tt.station.VMType == 1 {
				assert.Equal(t, "GoStation®", result.Contents.Body.Contents[1].(BoxTemplate).Contents[0].(TextTemplate).Text)
			} else {
				assert.Equal(t, "Super GoStation®", result.Contents.Body.Contents[1].(BoxTemplate).Contents[0].(TextTemplate).Text)
			}
			assert.Equal(t, tt.station.Address, result.Contents.Body.Contents[2].(BoxTemplate).Contents[0].(BoxTemplate).Contents[1].(TextTemplate).Text)
			assert.Equal(t, fmt.Sprintf("%.2f 公里", tt.station.Distance), result.Contents.Body.Contents[2].(BoxTemplate).Contents[1].(BoxTemplate).Contents[1].(TextTemplate).Text)
			assert.Equal(t, fmt.Sprintf("https://www.google.com.tw/maps/dir//%f,%f", tt.station.Latitude, tt.station.Longitude), result.Contents.Footer.Contents[0].(ButtonTemplate).Action.URI)
		})
	}
}
