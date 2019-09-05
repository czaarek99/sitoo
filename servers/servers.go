package servers

import (
	"encoding/json"
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

func getBadRequestResponse(text string) domain.ErrorResponse {
	return domain.ErrorResponse{
		ErrorText:    text,
		ResponseCode: 400,
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

type MultipleGetResponse struct {
	TotalCount uint32 `json:"totalCount"`
	Items      []domain.Product
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

	writer.Header().Set("Content-Type", "application/json")

	var errorResponse domain.ErrorResponse
	var successResponse interface{}

	path := request.URL.Path

	if strings.HasPrefix(path, "/api/products") {

		if request.Method == "GET" {

		} else if request.Method == "POST" {

		} else if request.Method == "PUT" {

		} else if request.Method == "DELETE" {

		}

	} else {
		errorResponse = domain.ErrorResponse{
			ErrorText:    "Not found",
			ResponseCode: 404,
		}
	}

	var err error
	var jsonBytes []byte

	if errorResponse.ResponseCode != 0 {
		jsonBytes, err = json.Marshal(errorResponse)
		writer.WriteHeader(errorResponse.ResponseCode)
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
