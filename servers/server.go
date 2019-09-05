package servers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"sitoo/domain"
	"strconv"
	"strings"
)

const (
	singleGET   = iota
	multipleGET = iota
)

type Server struct {
	Service domain.ProductService
}

type parsedGET struct {
	productID domain.ProductId
	getType   int

	start   uint64
	num     uint64
	sku     string
	barcode string
	fields  []string
}

type MultipleGetResponse struct {
	TotalCount uint32 `json:"totalCount"`
	Items      []domain.Product
}

func getBadRequestResponse(text string) domain.ErrorResponse {
	return domain.ErrorResponse{
		ErrorText:    text,
		ResponseCode: 400,
	}
}

func getNotFoundResponse() domain.ErrorResponse {
	return domain.ErrorResponse{
		ErrorText:    "Not found",
		ResponseCode: 404,
	}
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
		parsed.getType = singleGET
	} else {
		query := request.URL.Query()

		start, err := strconv.ParseUint(query["start"][0], 10, 64)

		if err != nil {
			parsed.start = start
		}

		num, numErr := strconv.ParseUint(query["num"][0], 10, 64)

		if numErr != nil {
			parsed.num = num
		}

		parsed.sku = query["sku"][0]
		parsed.barcode = query["barcode"][0]
		parsed.fields = strings.Split(query["fields"][0], ",")

		parsed.getType = multipleGET
	}

	return parsed
}

func (server Server) handleGET(
	request *http.Request,
) (interface{}, domain.ErrorResponse) {

	parsed := parseGET(request)

	if parsed.getType == singleGET {
		product, error := server.Service.GetProduct(parsed.productID, parsed.fields)

		if error != nil {
			return MultipleGetResponse{}, getBadRequestResponse(error.Error())
		} else {
			return product, domain.ErrorResponse{}
		}
	} else {
		products, count, error := server.Service.GetProducts(
			parsed.start,
			parsed.num,
			parsed.sku,
			parsed.barcode,
			parsed.fields,
		)

		if error != nil {
			return MultipleGetResponse{}, getBadRequestResponse(error.Error())
		} else {
			return MultipleGetResponse{
				TotalCount: count,
				Items:      products,
			}, domain.ErrorResponse{}
		}
	}
}

func (server Server) handlePOST(
	request *http.Request,
) (uint32, domain.ErrorResponse) {

	return 0, domain.ErrorResponse{}
}

func (server Server) handlePUT(
	request *http.Request,
) (bool, domain.ErrorResponse) {

	return false, domain.ErrorResponse{}
}

func (server Server) handleDELETE(
	request *http.Request,
) (bool, domain.ErrorResponse) {

	return false, domain.ErrorResponse{}
}

func (server Server) HandleRequest(
	writer http.ResponseWriter,
	request *http.Request,
) {

	var errorResponse domain.ErrorResponse
	var jsonResponse interface{}
	var stringResponse string

	path := request.URL.Path

	if strings.HasPrefix(path, "/api/products") {

		if request.Method == "GET" {
			jsonResponse, errorResponse = server.handleGET(request)
		} else if request.Method == "POST" {
			productID, err := server.handlePOST(request)

			errorResponse = err
			stringResponse = strconv.FormatUint(uint64(productID), 10)
		} else if request.Method == "PUT" {
			success, err := server.handlePUT(request)

			errorResponse = err
			stringResponse = strconv.FormatBool(success)
		} else if request.Method == "DELETE" {
			success, err := server.handleDELETE(request)

			errorResponse = err
			stringResponse = strconv.FormatBool(success)

		} else {
			errorResponse = getNotFoundResponse()
		}

	} else {
		errorResponse = getNotFoundResponse()
	}

	if stringResponse != "" {
		fmt.Fprint(writer, stringResponse)
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.Header().Set("Content-Type", "application/json")

		var jsonBytes []byte
		var err error

		if errorResponse.ResponseCode == 0 {
			writer.WriteHeader(http.StatusOK)

			jsonBytes, err = json.Marshal(jsonResponse)
		} else {
			jsonBytes, err = json.Marshal(errorResponse)
		}

		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
		} else {
			writer.WriteHeader(errorResponse.ResponseCode)
			writer.Write(jsonBytes)
		}
	}
}
