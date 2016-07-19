package main

// TODO
// - Add more unit tests - 100% coverage
// - Benchmark tests?
// - Error function
// - Logging
// - goroutines

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var log = logging.MustGetLogger("example")

const maxMessageSize int64 = 1048576

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
	Result float64   `json:"result"`
	Time   time.Time `json:"time"`
}

type ErrorResponse struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Time        time.Time `json:"time"`
}

var validPath = regexp.MustCompile("^/(calc|calc)$")

func calcHandler(w http.ResponseWriter, r *http.Request) {

	var body []byte = readMessage(r)

	log.Infof("%+v\n", string(body))

	var newCalc CalcRequest

	unmarshalCalcRequest(w, &body, &newCalc)

	var requiredButMissing []string

	for _, element := range calcRequestRequired {
		log.Debug(element)
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

		log.Warning(errorText)
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
		log.Error("ReadAll error")
		panic(err)
	}

	if err := r.Body.Close(); err != nil {
		log.Error("Body.Close() error")
		panic(err)
	}

	return body
}

// Unmarshal the calculation request
// TODO: Pass w by reference?
func unmarshalCalcRequest(w http.ResponseWriter, body *[]byte, newCalc *CalcRequest) {
	if err := json.Unmarshal(*body, newCalc); err != nil {
		// TODO: Move error JSON code to function.
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		//w.WriteHeader(422) // unprocessable entity
		errorJson := &ErrorResponse{
			Type:        "Unmarshal error",
			Description: err.Error(),
			Time:        time.Now(),
		}

		// TODO: Is MarshalIndent really needed? Maybe for logging only?
		prettyJson, err := json.MarshalIndent(errorJson, "", "\t")
		w.Write(prettyJson)
		if err != nil {
			// Only panic if unable to send an error response
			// Eventually replace this with error logging
			panic(err)
		}
	}
}

// Takes two floats and multiplies them,
func doCalculation(op1, op2 *float64) *SuccessResponse {
	return &SuccessResponse{Result: *op1 * *op2, Time: time.Now()}
}

// Decorator for handler functions
// Adds logging and validates paths
// TODO: Not sure if this is required with gorilla mux
func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info(r.URL.Path)
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/calc", makeHandler(calcHandler))
	return r
}

// Helper function to initialise the go-logging package.
func initLogging() {
	// Example format string. Everything except the message has a custom color
	// which is dependent on the log level. Many fields have a custom output
	// formatting too, eg. the time returns the hour down to the milli second.
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} - %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	stdout := logging.NewLogBackend(os.Stderr, "", 0)
	stdoutFormatted := logging.NewBackendFormatter(stdout, format)
	logging.SetBackend(stdoutFormatted)
}

func main() {
	initLogging()
	log.Notice("Calc server up...")
	http.ListenAndServe(":8080", Router())
}
