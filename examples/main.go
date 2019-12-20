package main

import (
	"errors"
	"github.com/imtanmoy/httpx"
	"log"
	"net/http"
	"net/url"
)

type Test struct {
	Name string `json:"name"`
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		var test Test
		test.Name = "Example Name"
		httpx.ResponseJSON(w, 200, &test)
	case "POST":
		var test Test
		err := httpx.DecodeJSON(r, &test)
		if err != nil {
			var mr *httpx.MalformedRequest
			if errors.As(err, &mr) {
				httpx.ResponseJSONError(w, r, mr.Status, mr.Status, mr.Msg)
			} else {
				httpx.ResponseJSONError(w, r, http.StatusInternalServerError, err)
			}
			return
		}
		httpx.ResponseJSON(w, 200, &test)
	case "PUT":
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"message": "put called"}`))
	case "DELETE":
		httpx.NoContent(w)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "not found"}`))
	}
}

func err(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		httpx.ResponseJSONError(w, r, 404, "custom error")
	case "POST":
		validationErrors := url.Values{
			"name": []string{"name is required"},
		}
		httpx.ResponseJSONError(w, r, 400, "Invalid Request", validationErrors)
	case "PUT":
		err := errors.New("this is an error")
		httpx.ResponseJSONError(w, r, 500, err)
	case "DELETE":
		httpx.NoContent(w)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "not found"}`))
	}
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/error", err)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
