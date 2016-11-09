package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

// Parse response JSON but don't test values returned, just print them
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
	responseBody := readBodyWithLimit(recorder.Result().Body)

	if err := json.Unmarshal(responseBody, &response); err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Printf("Result: %v\n", response.Result)
	fmt.Printf("Result: %v\n", response.Time)
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

	var response CalcResponse
	responseBody := readBodyWithLimit(recorder.Result().Body)

	// TODO: Create an unmarshalCalcResponse func that will error here, then handle it or fail
	if err := json.Unmarshal(responseBody, &response); err != nil {
		fmt.Printf("Failed to unmarshal response: %v\n", response.Result)
		t.Error(err)
		t.FailNow()
	}

	fmt.Println(response)

	fmt.Printf("Result: %v\n", response.Result)
	fmt.Printf("Result: %v\n", response.Time)

	//if err := json.Unmarshal(responseBody, &response); err != nil {
	//if serr, ok := err.(*json.SyntaxError); ok {
	//Info.Printf("Received expected json.SyntaxError: %v\n", serr)
	//} else {
	//t.Error("Unexpected json.Unmarshal error")
	//}
	//}
}
