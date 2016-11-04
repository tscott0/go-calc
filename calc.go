package main

// TODO
// - Add more unit tests - 100% coverage
// - Benchmark tests?
// - Error function
// - Logging
// - goroutines
// - Remove reflect package?

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	maxMessageSize int64 = 1048576
)

var (
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	LogFmt  int
)

// CalcRequest Use pointers here to distinguish nil values.
// If it's required then it must be non-nil
// TODO: Strictness in additional fields? Should we error if we
// receive unexpected fields, rather than silently discared them?
type CalcRequest struct {
	Operand1 *float64 `json:"operand1"`
	Operand2 *float64 `json:"operand2"`
}

// Variable names that are required in the CalcRequest
var calcRequestRequired = []string{"Operand1", "Operand2"}

type SuccessResponse struct {
	Result string    `json:"result"`
	Time   time.Time `json:"time"`
}

type ErrorResponse struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Time        time.Time `json:"time"`
}

var validPath = regexp.MustCompile("^/calc$")

func calcHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte = readMessage(r)
	var newCalc CalcRequest
	var requiredButMissing []string

	unmarshalCalcRequest(w, &body, &newCalc)

	for _, element := range calcRequestRequired {
		Debug.Print(element)
		v := reflect.ValueOf(newCalc)
		found := v.FieldByName(element)

		// TODO: Store the expected type of the required field too
		// it should fail to unmarshal but best to check anyway.
		if found.IsNil() {
			requiredButMissing = append(requiredButMissing, element)
		}
	}

	if len(requiredButMissing) > 0 {
		errorText := "The following elements are required but were not provided: "
		errorText += strings.Join(requiredButMissing, ", ")

		Warning.Print(errorText)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	result := doCalculation(newCalc.Operand1, newCalc.Operand2)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		panic(err)
	}
}

// Read the message from the http request.
// Limit the size to maxMessageSize (1MB)
// TODO: Tune this limit. 1MB is arbitrary
func readMessage(r *http.Request) []byte {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxMessageSize))

	if err != nil {
		Error.Print("ReadAll error")
		panic(err)
	}

	if err := r.Body.Close(); err != nil {
		Error.Print("Body.Close() error")
		panic(err)
	}

	return body
}

// Unmarshal the calculation request
func unmarshalCalcRequest(w http.ResponseWriter, body *[]byte, newCalc *CalcRequest) {
	if err := json.Unmarshal(*body, newCalc); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		errorJson := &ErrorResponse{
			Type:        "Unmarshal error",
			Description: err.Error(),
			Time:        time.Now(),
		}

		errResponse, err := json.Marshal(errorJson)
		if err != nil {
			Error.Print("Failed to build error json")
		}

		_, err = w.Write(errResponse)
		if err != nil {
			Error.Print("Failed to send error json")
		}
	}
}

// Takes two floats and multiplies them,
func doCalculation(op1, op2 *float64) *SuccessResponse {
	result := fmt.Sprintf("%f", *op1**op2)
	Info.Printf("%v * %v = %v\n", *op1, *op2, result)
	return &SuccessResponse{Result: result, Time: time.Now()}
}

// Decorator for handler functions
// Adds logging and validates paths
func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Debug.Print(r.URL.Path)
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

func main() {
	// Log line format
	//    2009/01/23 01:23:23.123123 d.go:23:
	LogFmt = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	//Debug = log.New(os.Stdout, "DEBUG: ", LogFmt)
	Debug = log.New(ioutil.Discard, "DEBUG: ", LogFmt)
	Info = log.New(os.Stdout, "INFO: ", LogFmt)
	Warning = log.New(os.Stdout, "WARNING: ", LogFmt)
	Error = log.New(os.Stderr, "ERROR: ", LogFmt)

	http.HandleFunc("/calc", makeHandler(calcHandler))
	Info.Print("Calc server up...")
	http.ListenAndServe(":8080", nil)
}
