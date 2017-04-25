package mixpanel

import (
	"os"
	"testing"

	"github.com/jipiboily/forwardlytics/integrations"
)

func TestIdentify(t *testing.T) {
	os.Setenv("MIXPANEL_TOKEN", "321")
	m := Mixpanel{}
	api := APIMock{Url: "http://www.example.com"}
	m.api = &api
	identification := integrations.Identification{
		UserID: "123",
		UserTraits: map[string]interface{}{
			"email": "john@example.com",
		},
		Timestamp:  1234567,
		ReceivedAt: 8765432,
	}
	err := m.Identify(identification)
	if err != nil {
		t.Fatal(err)
	}

	if api.Method != "GET" {
		t.Errorf("Expected method to be GET, was: %v", api.Method)
	}

	if api.Endpoint != "engage" {
		t.Errorf("Expected endpoint to be engage, was: %v", api.Endpoint)
	}

	expectedPayload := `{"$set":{"forwardlyticsReceivedAt":8765432,"forwardlyticsTimestamp":1234567},"$distinct_id":"123","$token":"321","$name":"","$email":"john@example.com"}`
	if string(api.Payload) != expectedPayload {
		t.Errorf("Expected payload: "+string(expectedPayload)+" got: %s", api.Payload)
	}
}

// Test track

// Test page

// Test request

type APIMock struct {
	Url      string
	Method   string
	Endpoint string
	Payload  []byte
}

func (api *APIMock) request(method string, endpoint string, payload []byte) error {
	api.Method = method
	api.Endpoint = endpoint
	api.Payload = payload
	return nil
}
