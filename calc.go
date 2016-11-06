package main

// TODO
// - Add more unit tests - 100% coverage
// - Benchmark tests?
// - Error function
// - goroutines
// - Remove reflect package?
// - Use struct tags (reflect.StructTag) to mark required fields
// - Return errors in functions and handle them in callers
// - Handle port conflicts

import (
	"encoding/json"
	"flag"
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
	logFmt  int
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

type CalcResponse struct {
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
	var body []byte = readBodyWithLimit(r.Body)
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
	w.WriteHeader(http.StatusOK)
	result := doCalculation(newCalc.Operand1, newCalc.Operand2)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		Error.Print("Failed to encode results")
		panic(err)
	}
}

// Read the message body from the http request.
// Limit the size to maxMessageSize (1MB)
// Closes the body and returns the unpacked bytes on success.
// TODO: Tune this limit. 1MB is arbitrary
func readBodyWithLimit(body io.ReadCloser) []byte {
	unpacked, err := ioutil.ReadAll(io.LimitReader(body, maxMessageSize))
	if err != nil {
		Error.Print("Failed to read from the HTTP Request.Body")
		panic(err)
	}

	if err := body.Close(); err != nil {
		Error.Print("Failed to call Close() on HTTP Request.Body")
		panic(err)
	}

	return unpacked
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
func doCalculation(op1, op2 *float64) *CalcResponse {
	result := fmt.Sprintf("%f", *op1**op2)
	Info.Printf("%v * %v = %v\n", *op1, *op2, result)
	return &CalcResponse{Result: result, Time: time.Now()}
}

// Decorator for debug logging and URL validation
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

func init() {

	// Log line format 2009/01/23 01:23:23.123123 d.go:23:
	logFmt = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	var debugMode = flag.Bool("debug", false, "Log debug messages to stdout")

	flag.Parse()

	Info = log.New(os.Stdout, "INFO: ", logFmt)
	Warning = log.New(os.Stdout, "WARNING: ", logFmt)
	Error = log.New(os.Stderr, "ERROR: ", logFmt)

	if *debugMode {
		Debug = log.New(os.Stdout, "DEBUG: ", logFmt)
		Info.Print("INIT: Debugging enabled")
	} else {
		Debug = log.New(ioutil.Discard, "DEBUG: ", logFmt)
		Info.Print("INIT: Debugging disabled")
	}

}

func main() {
	http.HandleFunc("/calc", makeHandler(calcHandler))
	Info.Print("INIT: Calc server up...")
	http.ListenAndServe(":8080", nil)
}
