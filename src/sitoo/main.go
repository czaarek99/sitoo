package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	ErrorText string `json:"errorText"`
}

func handler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	var response interface{}

	if strings.HasPrefix(request.URL.Path, "/api/products") {

		if request.Method == "GET" {

		} else if request.Method == "POST" {

		} else if request.Method == "PUT" {

		} else if request.Method == "DELETE" {

		}

	} else {
		response = ErrorResponse{
			ErrorText: "Route not found",
		}
	}

	jsonBytes, err := json.Marshal(response)

	if err == nil {
		writer.Write(jsonBytes)
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
	}

}

func main() {
	fmt.Println("Starting server")

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
