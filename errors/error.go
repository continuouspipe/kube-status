package errors

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"net/http"
)

type ApiError struct {
	Error string `json:"error"`
}

func LogAndRespondWithError(w http.ResponseWriter, code int, format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	glog.Error(err)
	glog.Flush()
	respondWithError(w, err.Error(), code)
}

//Returns the error in a json format
//if something wrong happens it will write the error on the response writer as a normal string
func respondWithError(w http.ResponseWriter, err string, code int) {
	apiError := ApiError{err}
	if res, err := json.Marshal(apiError); err != nil {
		http.Error(w, string(res), code)
	}
	http.Error(w, err, code)
}

type ErrorListProvider interface {
	error
	Add(elems ...error)
	Items() []error
}

type ErrorList struct {
	errors []error
}

func NewErrorList() *ErrorList {
	return &ErrorList{}
}

func (el *ErrorList) Add(elems ...error) {
	el.errors = append(el.errors, elems...)
}

func (el *ErrorList) AddErrorf(format string, a ...interface{}) {
	err := fmt.Errorf(format, a...)
	el.errors = append(el.errors, err)
}

func (el ErrorList) Items() []error {
	return el.errors
}

func (el ErrorList) Error() (err string) {
	if len(el.errors) <= 0 {
		return
	}

	err = fmt.Sprintf("An error occured: %s", el.errors[0].Error())
	if len(el.errors) == 1 {
		return
	}

	err = err + fmt.Sprintf("\nError stack:\n")

	for key, item := range el.errors {
		if key == 0 {
			continue
		}
		err = err + fmt.Sprintf("\n[%d] %s", key, item.Error())
	}
	return
}
