package api

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"net/http"
)

type ApiError struct {
	Error string `json:"error"`
}

func logErrorsAndRespondWithError(w http.ResponseWriter, code int, errList ErrorList) {
	err := errList.String()
	glog.Error(err)
	glog.Flush()
	respondWithError(w, err, code)
}

func logAndRespondWithError(w http.ResponseWriter, code int, format string, args ...interface{}) {
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

type ErrorList struct {
	Items []error
}

func (el *ErrorList) Add(err error) {
	el.Items = append(el.Items, err)
}

func (el ErrorList) HasErrors() bool {
	return len(el.Items) > 0
}

func (el ErrorList) String() (err string) {

	if len(el.Items) <= 0 {
		return
	}

	err = fmt.Sprintf("An error occured: %s", el.Items[0].Error())
	if len(el.Items) == 1 {
		return
	}

	err = err + fmt.Sprintf("\nError stack:\n")

	for key, error := range el.Items {
		err = err + fmt.Sprintf("\n[%d] %s", key, error.Error())
	}
	return
}
