package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jipiboily/forwardlytics/integrations"
)

func TestIdentifyWhenNotPOST(t *testing.T) {
	expectedStatusCode := 404
	expectedBody := "404 page not found"

	r, err := http.NewRequest("GET", "/track", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	Identify(w, r)

	if w.Code != expectedStatusCode {
		t.Errorf("Wrong status code. Expecting %v but got %v", expectedStatusCode, w.Code)
	}

	if !strings.Contains(w.Body.String(), expectedBody) {
		t.Errorf(`Wrong response. Expecting "%s" but got "%s"`, expectedBody, w.Body.String())
	}
}

func TestIdentifyWhenInvalidJSON(t *testing.T) {
	expectedStatusCode := 400
	expectedBody := `{"message": "Invalid request."}`

	requestBody := `invalid JSON here`
	r, err := http.NewRequest("POST", "/track", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	Identify(w, r)

	if w.Code != expectedStatusCode {
		t.Errorf("Wrong status code. Expecting %v but got %v", expectedStatusCode, w.Code)
	}

	if !strings.Contains(w.Body.String(), expectedBody) {
		t.Errorf(`Wrong response. Expecting "%s" but got "%s"`, expectedBody, w.Body.String())
	}
}

func TestIdentifyWhenMissingParameter(t *testing.T) {
	expectedStatusCode := 400
	expectedBody := `{"message": "Missing parameters: userID, timestamp."}`

	requestBody := `{}`
	r, err := http.NewRequest("POST", "/track", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	Identify(w, r)

	if w.Code != expectedStatusCode {
		t.Errorf("Wrong status code. Expecting %v but got %v", expectedStatusCode, w.Code)
	}

	if !strings.Contains(w.Body.String(), expectedBody) {
		t.Errorf(`Wrong response. Expecting "%s" but got "%s"`, expectedBody, w.Body.String())
	}
}

func TestIdentifyWhenOneIntegrationFails(t *testing.T) {
	expectedStatusCode := 500
	expectedBody := `{"message": "Fatal error during identification with an integration (test-only-integration-failing): some random error"}`

	requestBody := `{
		"name":"something.created",
		"userID":"123",
		"properties": { "someCounter": 97 },
		"timestamp": 12345678
	}`
	r, err := http.NewRequest("POST", "/track", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	failingIntegration := FailingIntegrationIdentification{}
	integrations.RegisterIntegration("test-only-integration-failing", failingIntegration)
	defer integrations.RemoveIntegration("test-only-integration-failing")

	workingIntegration := FakeIntegration{}
	integrations.RegisterIntegration("test-only-integration-working", workingIntegration)
	defer integrations.RemoveIntegration("test-only-integration-working")

	Identify(w, r)

	if w.Code != expectedStatusCode {
		t.Errorf("Wrong status code. Expecting %v but got %v", expectedStatusCode, w.Code)
	}

	if !strings.Contains(w.Body.String(), expectedBody) {
		t.Errorf(`Wrong response. Expecting "%s" but got "%s"`, expectedBody, w.Body.String())
	}
}

func TestIdentifyWhenValid(t *testing.T) {
	expectedStatusCode := 200
	expectedBody := `{"message": "Forwarding identify to integrations."}`

	requestBody := `{
		"name":"something.created",
		"userID":"123",
		"properties": { "someCounter": 97 },
		"timestamp": 12345678
	}`
	r, err := http.NewRequest("POST", "/track", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	integration := &CalledIntegrationIdentify{t: t}
	integrations.RegisterIntegration("test-only-integration-called", integration)
	defer integrations.RemoveIntegration("test-only-integration-called")

	Identify(w, r)

	if !integration.Identified {
		t.Error("Identify was not called on the integration")
	}

	if w.Code != expectedStatusCode {
		t.Errorf("Wrong status code. Expecting %v but got %v", expectedStatusCode, w.Code)
	}

	if !strings.Contains(w.Body.String(), expectedBody) {
		t.Errorf(`Wrong response. Expecting "%s" but got "%s"`, expectedBody, w.Body.String())
	}
}

// FailingIntegrationIdentification is an integration that fails when called
type FailingIntegrationIdentification struct {
	FakeIntegration
}

// Identify is failing in this case
func (fi FailingIntegrationIdentification) Identify(identification integrations.Identification) error {
	return errors.New("some random error")
}

// Enabled returns true because this failing integraiton is enabled
func (FailingIntegrationIdentification) Enabled() bool {
	return true
}

type CalledIntegrationIdentify struct {
	FakeIntegration
	Identified bool
	t          *testing.T
}

func (i *CalledIntegrationIdentify) Identify(identification integrations.Identification) error {
	i.Identified = true

	expectedReceivedAtCloseTo := time.Now().Unix() - 5
	if identification.ReceivedAt < expectedReceivedAtCloseTo {
		i.t.Errorf("ReceivedAt looks wrong. Expecting something close to %v but got %v", expectedReceivedAtCloseTo, identification.ReceivedAt)
	}

	return nil
}

func (i CalledIntegrationIdentify) Enabled() bool {
	return true
}
