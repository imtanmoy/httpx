package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var ErrInternalServerError = errors.New("internal server error")

// JSONError construct http error response
type JSONError struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	Code      int        `json:"code"`    // application-specific error code
	Message   string     `json:"message"` // user-level status message
	Errors    url.Values `json:"errors"`  // form input field errors
	Timestamp time.Time  `json:"timestamp"`
	Path      string     `json:"path"`
}

func (j *JSONError) setPath(path string) {
	j.Path = path
}

func (j *JSONError) setHTTPStatusCode(statusCode int) {
	j.HTTPStatusCode = statusCode
}

func (j *JSONError) Error() string {
	message := ""
	if j.Message != "" {
		message = j.Message
	} else if j.Message == "" && j.Err != nil {
		message = j.Err.Error()
	} else if j.Message == "" && j.Err == nil {
		message = ErrInternalServerError.Error()
	}
	return fmt.Sprintf("%d: %s", j.Code, message)
}

//Can easily output properly formatted JSON error messages for REST API services.
func (j *JSONError) Render() map[string]interface{} {

	response := make(map[string]interface{})

	if j.Code == 0 {
		if j.HTTPStatusCode != 0 {
			response["code"] = j.HTTPStatusCode
		} else {
			response["code"] = http.StatusInternalServerError
		}
	} else {
		response["code"] = j.Code
	}

	if j.Message != "" {
		response["message"] = j.Message
	} else if j.Message == "" && j.Err != nil {
		response["message"] = j.Err.Error()
	} else if j.Message == "" && j.Err == nil {
		response["message"] = "Oops! Something went wrong"
	}
	if j.Errors == nil {
		response["errors"] = make(url.Values)
	}
	response["path"] = j.Path
	return response
}

// NewJSONError create new JSONError struct
func NewJSONError(statusCode int, err error, code int, message string, errs url.Values, path string) *JSONError {
	if statusCode < 400 || statusCode >= 600 {
		statusCode = 500
	}
	if errs == nil {
		errs = make(url.Values)
	}
	return &JSONError{
		Err:            err,
		HTTPStatusCode: statusCode,
		Code:           code,
		Message:        message,
		Errors:         errs,
		Timestamp:      time.Now().UTC(),
		Path:           path,
	}
}

func parseJSONError(value ...interface{}) *JSONError {
	var err error
	var code int
	var message string
	var errs url.Values

	if len(value) == 0 {
		err = ErrInternalServerError
	}
	for i, val := range value {
		if i >= 4 {
			break
		}
		switch v := val.(type) {
		case int:
			code = v
			break
		case string:
			message = v
			break
		case error:
			err = v
		case url.Values:
			errs = v
			break
		}
	}
	if code == 0 {
		code = http.StatusInternalServerError
	}
	if errs == nil {
		errs = make(url.Values)
	}
	if err == nil {
		err = ErrInternalServerError
	}
	if message == "" {
		if errors.Is(err, ErrInternalServerError) {
			message = "Oops! Something went wrong"
		} else {
			message = err.Error()
		}
	}
	return &JSONError{
		Err:       err,
		Code:      code,
		Message:   message,
		Errors:    errs,
		Timestamp: time.Now().UTC(),
	}
}

// ResponseJSONError create new JSONError struct
func ResponseJSONError(w http.ResponseWriter, r *http.Request, status int, value ...interface{}) {
	jsonErr := parseJSONError(value)
	jsonErr.setPath(r.RequestURI)
	jsonErr.setHTTPStatusCode(status)
	ResponseJSON(w, status, jsonErr)
}
