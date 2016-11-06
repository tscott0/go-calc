package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Unmarshal the calculation response. Added this function to support testing.
func unmarshalCalcResponse(body *[]byte, response *CalcResponse) {
	if err := json.Unmarshal(*body, response); err != nil {
		// Could indicate a malformed response or an ErrorResponse was returned
		Error.Print("Failed to unmarshal response")
	}
}

// Test for HTTP 200 status on success
func TestStatusOK(t *testing.T) {
	userJson := `{"operand1": 1.4, "operand2": 2.3}`

	req, err := http.NewRequest("POST", "/calc", strings.NewReader(userJson))
	if err != nil {
		log.Fatal("Could not build test request")
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(makeHandler(calcHandler))

	handler.ServeHTTP(recorder, req)

	// Check for status 200
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Unexpected status code: received %v expected %v",
			status, http.StatusOK)
	}
}

// Test for HTTP 404 status using invalid URL
func TestStatusNotFound(t *testing.T) {
	userJson := `{"operand1": 1.4, "operand2": 2.3}`

	req, err := http.NewRequest("POST", "/404path", strings.NewReader(userJson))

	if err != nil {
		log.Fatal("Could not build test request")
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(makeHandler(calcHandler))

	handler.ServeHTTP(recorder, req)

	// Check for status 200
	if status := recorder.Code; status != http.StatusNotFound {
		t.Errorf("Unexpected status code: received %v expected %v",
			status, http.StatusNotFound)
	}
}

// Parse response and check correct value
func TestBasicCalc(t *testing.T) {
	userJson := `{"operand1": 1.4, "operand2": 2.3}`

	req, err := http.NewRequest("POST", "/calc", strings.NewReader(userJson))
	if err != nil {
		log.Fatal("Could not build test request")
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(makeHandler(calcHandler))

	handler.ServeHTTP(recorder, req)

	var response CalcResponse
	unmarshalCalcResponse(&readMessage(recorder), &response)

}
