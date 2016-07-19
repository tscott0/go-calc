package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TODO: Refactor tests to remove boilerplate
// Teardown server for each request? Probably not needed
func TestSuccess(t *testing.T) {
	var request = []byte(`{"operand1": 1.4, "operand2": 2.3}`)

	r, err := http.NewRequest("POST", "/calc", bytes.NewBuffer(request))
	if err != nil {
		log.Fatal("Could not build test request")
	}
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	Router().ServeHTTP(w, r)

	var success SuccessResponse

	if err := json.Unmarshal(w.Body.Bytes(), &success); err != nil {
		t.Error("Failed to unmarshal successful response")
	}

	// TODO: Assertions

	fmt.Println("Result: ", success.Result)

}
