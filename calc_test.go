package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Unmarshal the errpr response
func unmarshalErrorResponse(body *[]byte, resp *ErrorResponse, t *testing.T) error {
	if err := json.Unmarshal(*body, resp); err != nil {
		return &UnmarshalError{"Failed to parse JSON"}
	}

	// error and time must be non-zero in the error response
	if resp.Error == "" {
		return &UnmarshalError{"error was empty"}
	} else if resp.Time.IsZero() {
		return &UnmarshalError{"time was zero"}
	}

	return nil
}

func TestFlags(t *testing.T) {
	testCases := []struct {
		name       string
		silentMode bool
		debugMode  bool
	}{
		{"Flags:<none>", false, false},
		{"Flags:--silent", true, false},
		{"Flags:--debug", false, true},
		{"Flags:--silent,--debug", true, true},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			initLogging(tc.silentMode, tc.debugMode)
		})
	}

	// Return logging to default state
	initLogging(*silent, *debug)
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

// Run main() and stop after a second
// Port must not be in use
func TestMain(t *testing.T) {
	go main()
	time.Sleep(time.Second)
	return
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

// Parse response JSON and verify result
// See TestSuccess for end-to-end testing of calculations
func TestSuccessResponseOnly(t *testing.T) {
	userJson := `{"operand1": 1.4, "operand2": 2.3}`
	expected := 3.22

	req, err := http.NewRequest("POST", "/calc", strings.NewReader(userJson))
	if err != nil {
		log.Fatal("Could not build test request")
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(makeHandler(calcHandler))

	handler.ServeHTTP(recorder, req)

	var response CalcResponse
	responseBody := readBodyWithLimit(recorder.Result().Body)

	if err := json.Unmarshal(responseBody, &response); err != nil {
		t.Error(err)
	}

	// Convert Result string to a 64-bit float for numeric comparison
	res, err := strconv.ParseFloat(response.Result, 64)
	if err != nil {
		t.Error(err)
	}

	if res != expected {
		t.Errorf("Result was %v, expected 3.22", response.Result)
	}
}

// Table-driven testing of successful requests.
// Results are in strings to check decimal places (6 d.p.)
func TestSuccess(t *testing.T) {
	testCases := []struct {
		name     string
		reqJson  string
		expected string
	}{
		{"1.4*2.3=3.220000", `{"operand1": 1.4, "operand2": 2.3}`, "3.220000"},
		{"8.9*7.6=3.220000", `{"operand1": 8.9, "operand2": 2.3}`, "20.470000"},
		{"0.1*0.0=0.000000", `{"operand1": 0.1, "operand2": 0.0}`, "0.000000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			req, err := http.NewRequest("POST", "/calc", strings.NewReader(tc.reqJson))
			if err != nil {
				log.Fatal("Could not build test request")
			}

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(makeHandler(calcHandler))

			handler.ServeHTTP(recorder, req)

			var response CalcResponse
			responseBody := readBodyWithLimit(recorder.Result().Body)

			if err := json.Unmarshal(responseBody, &response); err != nil {
				t.Error(err)
			}

			if strings.Compare(response.Result, tc.expected) != 0 {
				t.Errorf("Result was %v, expected %v", response.Result, tc.expected)
			}
		})
	}
}

// Unmarshal error from malformed JSON
func TestUnmarshalError(t *testing.T) {
	userJson := `malformed JSON example`

	req, err := http.NewRequest("POST", "/calc", strings.NewReader(userJson))
	if err != nil {
		log.Fatal("Could not build test request")
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(makeHandler(calcHandler))

	handler.ServeHTTP(recorder, req)

	responseBody := readBodyWithLimit(recorder.Result().Body)
	var errorResp ErrorResponse

	if err := unmarshalErrorResponse(&responseBody, &errorResp, t); err != nil {
		if ue, ok := err.(*UnmarshalError); ok {
			Info.Println(ue.Error())
			t.Fail()
		} else if mfe, ok := err.(*MissingFieldError); ok {
			Info.Println(mfe.Error())
			t.Fail()
		}
	}

	Info.Printf("Error response message: %v\n", errorResp.Error)
	Info.Printf("Error response time: %v\n", errorResp.Time)
}

// One required field missing in the request
func TestMissingFieldError(t *testing.T) {
	userJson := `{"operand1": 1.4}`

	req, err := http.NewRequest("POST", "/calc", strings.NewReader(userJson))
	if err != nil {
		log.Fatal("Could not build test request")
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(makeHandler(calcHandler))

	handler.ServeHTTP(recorder, req)

	responseBody := readBodyWithLimit(recorder.Result().Body)
	var errorResp ErrorResponse
	if err := json.Unmarshal(responseBody, &errorResp); err != nil {
		Info.Println("Failed to parse JSON")
		t.Fail()
	}

	Info.Printf("Error response message: %v\n", errorResp.Error)
	Info.Printf("Error response time: %v\n", errorResp.Time)

	// Finally check the actual error message
	if errorResp.Error != fmt.Sprintf("The following %v field(s) were missing: %v",
		1, "Operand2") {
		Info.Printf("Unexpected error response. Received \"%v\"", errorResp.Error)
		t.Fail()
	}

}

// TODO: Table-driven tests for errors, with subtests.
