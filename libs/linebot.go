package libs

import "fmt"

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

type WebhookEvent struct {
	Type    string `json:"type"`
	Message struct {
		Type            string  `json:"type"`
		Id              string  `json:"id"`
		Latitude        float64 `json:"latitude,omitempty"`
		Longitude       float64 `json:"longitude,omitempty"`
		Address         string  `json:"address,omitempty"`
		QuotedMessageId string  `json:"quotedMessageId,omitempty"`
		QuoteToken      string  `json:"quoteToken"`
		Text            string  `json:"text,omitempty"`
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

// Flex message template
type LayoutType string
type ButtonType string
type ActionType string
type ElementType string

const (
	// box layout
	BaselineLayout   LayoutType = "baseline"
	VerticalLayout   LayoutType = "vertical"
	HorizontalLayout LayoutType = "horizontal"
	// button style
	LinkButton      ButtonType = "link"
	PrimaryButton   ButtonType = "primary"
	SecondaryButton ButtonType = "secondary"
	// action type
	PostbackAction       ActionType = "postback"
	URIAction            ActionType = "uri"
	MessageAction        ActionType = "message"
	DatetimePickerAction ActionType = "datetimepicker"
	LocationAction       ActionType = "location"
	// flex element
	BoxElement       ElementType = "box"
	ImageElement     ElementType = "image"
	TextElement      ElementType = "text"
	ButtonElement    ElementType = "button"
	FilterElement    ElementType = "filter"
	SeparatorElement ElementType = "separator"
)

type BoxTemplate struct {
	Type     ElementType   `json:"type"`
	Layout   LayoutType    `json:"layout"`
	Margin   string        `json:"margin,omitempty"`
	Spacing  string        `json:"spacing,omitempty"`
	Contents []interface{} `json:"contents"`
}

type TextTemplate struct {
	Type   ElementType `json:"type"`
	Text   string      `json:"text"`
	Color  string      `json:"color,omitempty"`
	Size   string      `json:"size,omitempty"`
	Margin string      `json:"margin,omitempty"`
	Flex   int         `json:"flex"`
	Wrap   bool        `json:"wrap"`
	Weight string      `json:"weight,omitempty"`
}

type ActionTemplate struct {
	Type  ActionType `json:"type"`
	Label string     `json:"label"`
	URI   string     `json:"uri,omitempty"`
	Text  string     `json:"text,omitempty"`
}

type ButtonTemplate struct {
	Type   ElementType    `json:"type"`
	Style  ButtonType     `json:"style"`
	Height string         `json:"height,omitempty"`
	Action ActionTemplate `json:"action"`
}

type ContentTemplate struct {
	Type     ElementType   `json:"type"`
	Layout   LayoutType    `json:"layout"`
	Contents []interface{} `json:"contents"`
	Flex     int           `json:"flex"`
	Spacing  string        `json:"spacing,omitempty"`
}

type BubbleTemplate struct {
	Type   string          `json:"type"`
	Body   ContentTemplate `json:"body"`
	Footer ContentTemplate `json:"footer"`
}

type QuickReplyItemTemplate struct {
	Type     string         `json:"type"`
	ImageUrl string         `json:"imageUrl"`
	Action   ActionTemplate `json:"action"`
}

type BubbleMessageTemplate struct {
	Type     string         `json:"type"`
	AltText  string         `json:"altText"`
	Contents BubbleTemplate `json:"contents"`
}

type QuickReplyTemplate struct {
	Items []QuickReplyItemTemplate `json:"items"`
}

func BubbleMessage(station GoStation) BubbleMessageTemplate {
	message := BubbleMessageTemplate{
		Type:    "flex",
		AltText: "sogorro",
		Contents: BubbleTemplate{
			Type: "bubble",
			Body: ContentTemplate{
				Type:   BoxElement,
				Layout: VerticalLayout,
			},
			Footer: ContentTemplate{
				Type:    BoxElement,
				Layout:  VerticalLayout,
				Spacing: "sm",
				Flex:    0,
			},
		},
	}

	message.Contents.Body.Contents = append(message.Contents.Body.Contents, TextTemplate{
		Type:   "text",
		Text:   station.Location,
		Weight: "bold",
		Size:   "xl",
		Wrap:   true,
	})

	stationType := "GoStation®"
	if station.VMType == 3 {
		stationType = "Super GoStation®"
	}

	message.Contents.Body.Contents = append(message.Contents.Body.Contents, BoxTemplate{
		Type:   "box",
		Layout: BaselineLayout,
		Margin: "md",
		Contents: []interface{}{
			TextTemplate{
				Type:   TextElement,
				Text:   stationType,
				Size:   "sm",
				Color:  "#999999",
				Margin: "md",
				Flex:   0,
			},
		},
	})

	message.Contents.Body.Contents = append(message.Contents.Body.Contents, BoxTemplate{
		Type:    BoxElement,
		Layout:  VerticalLayout,
		Margin:  "lg",
		Spacing: "sm",
		Contents: []interface{}{
			BoxTemplate{
				Type:    BoxElement,
				Layout:  BaselineLayout,
				Spacing: "sm",
				Contents: []interface{}{
					TextTemplate{
						Type:  TextElement,
						Text:  "地址",
						Color: "#aaaaaa",
						Size:  "sm",
						Flex:  1,
					},
					TextTemplate{
						Type:  TextElement,
						Text:  station.Address,
						Wrap:  true,
						Color: "#666666",
						Size:  "sm",
						Flex:  5,
					},
				},
			},
			BoxTemplate{
				Type:    BoxElement,
				Layout:  BaselineLayout,
				Spacing: "sm",
				Contents: []interface{}{
					TextTemplate{
						Type:  TextElement,
						Text:  "距離",
						Color: "#aaaaaa",
						Size:  "sm",
						Flex:  1,
					},
					TextTemplate{
						Type:  TextElement,
						Text:  fmt.Sprintf("%.2f 公里", station.Distance),
						Color: "#666666",
						Size:  "sm",
						Flex:  5,
					},
				},
			},
		},
	})

	message.Contents.Footer.Contents = append(message.Contents.Footer.Contents, ButtonTemplate{
		Type:   ButtonElement,
		Style:  PrimaryButton,
		Height: "sm",
		Action: ActionTemplate{
			Type:  URIAction,
			Label: "立即前往",
			URI:   fmt.Sprintf("https://www.google.com.tw/maps/dir//%f,%f", station.Latitude, station.Longitude),
		},
	})

	return message
}

func WelcomeQuickReplyMessage() QuickReplyTemplate {
	message := QuickReplyTemplate{}
	message.Items = append(message.Items, QuickReplyItemTemplate{
		Type: "action",
		Action: ActionTemplate{
			Type:  LocationAction,
			Label: "分享位置",
		},
	})

	return message
}
