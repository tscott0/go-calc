package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

// Use pointers here to distinguish nil values.
// If it's required then it must be non-nil
// TODO: Refactor to build support for required fields
// Could be done by creating a slice of field names and calling
// FieldByName from the reflect package. If it's in the slice and
// in the struct then it must be non-nil
// TODO: Strictness in additional fields? Should we error if we
// receive unexpected fields, rather than silently discared them?
type CalcRequest struct {
	Operand1 *float64 `json:"operand1"`
	Operand2 *float64 `json:"operand2"`
}

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
	var newCalc CalcRequest

	// Read the body of the message, limited to 1MB
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", string(body))

	if err := json.Unmarshal(body, &newCalc); err != nil {

		// TODO: Move error JSON code to function.
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		//w.WriteHeader(422) // unprocessable entity
		errorJson := &ErrorResponse{Type: "Unmarshal error", Description: err.Error(), Time: time.Now()}

		// TODO: Is MarshalIndent really needed? Maybe for logging only?
		prettyJson, _ := json.MarshalIndent(errorJson, "", "\t")
		w.Write(prettyJson)
		panic(err)
	}

	fmt.Println(newCalc)

	// TODO: Tidy up this section.
	if newCalc.Operand1 == nil {
		fmt.Println("Operand1 not provided")
		return
	}

	if newCalc.Operand2 == nil {
		fmt.Println("Operand2 not provided")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	result := doCalculation(newCalc.Operand1, newCalc.Operand2)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		panic(err)
	}
}

// Takes two floats and multiplies them,
// TODO: Refactor to pass request and response JSON by ref and return err
func doCalculation(op1, op2 *float64) *SuccessResponse {
	return &SuccessResponse{Result: *op1 * *op2, Time: time.Now()}
}

// Decorator for handler functions
// Adds logging and validates paths
// TODO: Not sure if it's required with gorilla mux
func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/calc", makeHandler(calcHandler))

	http.ListenAndServe(":8080", r)
}
