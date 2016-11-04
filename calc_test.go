package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TODO: Refactor tests to remove boilerplate
// Teardown server for each request? Probably not needed
func TestBasic(t *testing.T) {

	reqBytes := []byte(`{"operand1": 1.4, "operand2": 2.3}`)

	req, err := http.NewRequest("POST", "/calc", bytes.NewBuffer(reqBytes))
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

	// TODO: Assertions

	//fmt.Println("Result: ", uccess.Result)

}
