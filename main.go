package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"path"
	"sitoo/domain"
	"sitoo/repositories"
	"sitoo/services"
	"strconv"
	"strings"
)

type ErrorResponse struct {
	ErrorText    string `json:"errorText"`
	responseCode int
}

type Server struct {
	service domain.ProductService
}

func getBadRequestResponse() ErrorResponse {
	return ErrorResponse{
		ErrorText:    "Bad request",
		responseCode: 400,
	}
}

const (
	single   = iota
	multiple = iota
)

type parsedGET struct {
	productID domain.ProductId
	query     map[string]string
	fields    []string
	getType   int
}

//Assumes properly formatted GET
//Too much time to write a parser
//Could use a library but this show i know more right?
func parseGET(request *http.Request) parsedGET {

	parsed := parsedGET{}

	base := path.Base(request.URL.Path)

	productID, err := strconv.ParseUint(base, 10, 32)

	if err == nil {
		parsed.productID = domain.ProductId(productID)
		parsed.getType = single
	}

	query := make(map[string]string)

	for key, values := range request.URL.Query() {
		value := values[0]

		if key == "fields" {
			parsed.fields = strings.Split(value, ",")
		} else {
			query[key] = value
		}
	}

	parsed.query = query
	parsed.getType = multiple

	return parsed
}

func handler(
	service domain.ProductService,
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Content-Type", "application/json")

	var errorResponse ErrorResponse
	var successResponse interface{}

	path := request.URL.Path

	if strings.HasPrefix(path, "/api/products") {

		if request.Method == "GET" {
			parsed := parseGET(request)

			if parsed.getType == single {
				successResponse, _ = service.GetProduct(parsed.productID, parsed.fields)
			}

		} else if request.Method == "POST" {

		} else if request.Method == "PUT" {

		} else if request.Method == "DELETE" {

		}

	} else {
		errorResponse = ErrorResponse{
			ErrorText:    "Not found",
			responseCode: 404,
		}
	}

	var err error
	var jsonBytes []byte

	if errorResponse.responseCode != 0 {
		jsonBytes, err = json.Marshal(errorResponse)
		writer.WriteHeader(errorResponse.responseCode)
	} else {
		jsonBytes, err = json.Marshal(successResponse)
	}

	if err == nil {
		writer.WriteHeader(http.StatusOK)
		writer.Write(jsonBytes)
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
	}

}

func main() {
	log.Println("Starting server")

	connection, err := sql.Open("mysql", "root:@/sitoo_test_assignment")

	if err != nil {
		log.Fatal("Could not connect to database")
	}

	repo := repositories.Repository{
		DB: connection,
	}

	service := services.Service{
		Repo: repo,
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		handler(service, writer, request)
	})

	http.ListenAndServe(":8080", nil)
}
