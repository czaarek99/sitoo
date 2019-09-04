package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	ErrorText string
}

func handler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	var response interface{}

	if strings.HasPrefix(request.URL.Path, "/api/products") {

	} else {
		response = ErrorResponse{
			ErrorText: "Route not found",
		}
	}

	jsonBytes, err := json.Marshal(response)

	if err == nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.Write(jsonBytes)
	}

}

func main() {
	fmt.Println("Starting server")

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
