package httperr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type Error struct {
	Status  int    `json:"-"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

var _ error = Error{}

var stdErrs = map[int]Error{}

// NewStd should only be called in "init" context (func init() or global declarations)
func NewStd(code, status int, message string) (err Error) {
	if prev, ok := stdErrs[code]; ok {
		panic(fmt.Errorf("error code already taken: %d (previous: %d %q)", code, prev.Status, prev.Message))
	}

	err = Error{status, code, message}
	stdErrs[code] = err

	return
}

func AllStd() (errs []Error) {
	keys := make([]int, 0, len(stdErrs))
	for i := range stdErrs {
		keys = append(keys, i)
	}
	sort.Ints(keys)

	errs = make([]Error, len(stdErrs))
	for i, k := range keys {
		errs[i] = stdErrs[k]
	}

	return
}

func New(status int, err error) Error {
	return Error{Status: status, Message: err.Error()}
}

func (err Error) Error() string {
	return err.Message
}

func (err Error) Any() bool {
	return err != Error{}
}

func (err Error) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}
