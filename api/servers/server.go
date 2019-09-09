package servers

import (
	"api/domain"
	"encoding/json"
	"net/http"
	"path"
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

type errorResponse struct {
	ErrorText    string `json:"errorText"`
	responseCode int
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

func getBadRequestResponse(text string) errorResponse {
	return errorResponse{
		ErrorText:    text,
		responseCode: 400,
	}
}

func getNotFoundResponse() errorResponse {
	return errorResponse{
		ErrorText:    "Not found",
		responseCode: 404,
	}
}

func writeError(
	writer http.ResponseWriter,
	errorResponse errorResponse,
) {
	writeJSON(writer, errorResponse, errorResponse.responseCode)
}

func writeJSON(
	writer http.ResponseWriter,
	item interface{},
	statusCode int,
) {
	jsonBytes, err := json.Marshal(item)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(statusCode)
		writer.Write(jsonBytes)
	}
}

func getProductIDFromPath(
	requestPath string,
) (domain.ProductId, error) {
	base := path.Base(requestPath)

	id, err := strconv.ParseUint(base, 10, 32)

	return domain.ProductId(id), err
}

//Assumes properly formatted GET
//Too much time to write a parser
//Could use a library but this show i know more right?
func parseGET(request *http.Request) parsedGET {

	parsed := parsedGET{}

	productID, err := getProductIDFromPath(request.URL.Path)

	query := request.URL.Query()

	if err == nil {
		parsed.productID = domain.ProductId(productID)
		parsed.getType = singleGET
	} else {
		start, err := strconv.ParseUint(query.Get("start"), 10, 64)

		if err == nil {
			parsed.start = start
		}

		num, numErr := strconv.ParseUint(query.Get("num"), 10, 64)

		if numErr == nil {
			parsed.num = num
		}

		parsed.sku = query.Get("sku")
		parsed.barcode = query.Get("barcode")
		parsed.getType = multipleGET
	}

	delimitedFields := query.Get("fields")

	if delimitedFields != "" {
		parsed.fields = strings.Split(delimitedFields, ",")
	}

	return parsed
}

func (server Server) handleGET(
	writer http.ResponseWriter,
	request *http.Request,
) {

	parsed := parseGET(request)

	if parsed.getType == singleGET {
		product, error := server.Service.GetProduct(parsed.productID, parsed.fields)

		if error != nil {
			writeError(writer, getBadRequestResponse(error.Error()))
		} else {
			writeJSON(writer, product, http.StatusOK)
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
			writeError(writer, getBadRequestResponse(error.Error()))
		} else {
			envelope := struct {
				TotalCount uint32           `json:"totalCount"`
				Items      []domain.Product `json:"items"`
			}{
				TotalCount: count,
				Items:      products,
			}

			writeJSON(writer, envelope, http.StatusOK)
		}
	}
}

func (server Server) handlePOST(
	writer http.ResponseWriter,
	request *http.Request,
) {

	var product domain.ProductAddInput

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&product)

	defer request.Body.Close()

	if err != nil {
		writeError(writer, getBadRequestResponse(err.Error()))
		return
	}

	var id domain.ProductId
	id, err = server.Service.AddProduct(product)

	if err != nil {
		writeError(writer, getBadRequestResponse(err.Error()))
		return
	}

	idString := strconv.FormatUint(uint64(id), 10)
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(idString))
}

func (server Server) handlePUT(
	writer http.ResponseWriter,
	request *http.Request,
) {

	id, err := getProductIDFromPath(request.URL.Path)

	if err != nil {
		writeError(writer, getBadRequestResponse("Missing product id to patch"))
		return
	}

	var changes domain.ProductUpdateInput

	decoder := json.NewDecoder(request.Body)
	err = decoder.Decode(&changes)

	defer request.Body.Close()

	if err != nil {
		writeError(writer, getBadRequestResponse(err.Error()))
		return
	}

	err = server.Service.UpdateProduct(id, changes)

	if err != nil {
		writeError(writer, getBadRequestResponse(err.Error()))
	} else {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("true"))
	}

}

func (server Server) handleDELETE(
	writer http.ResponseWriter,
	request *http.Request,
) {

	id, err := getProductIDFromPath(request.URL.Path)

	if err != nil {
		writeError(writer, getBadRequestResponse("Bad Request"))
		return
	}

	err = server.Service.DeleteProduct(id)

	if err != nil {
		writeError(writer, getBadRequestResponse(err.Error()))
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("true"))

}

func (server Server) HandleRequest(
	writer http.ResponseWriter,
	request *http.Request,
) {

	var notFoundError errorResponse

	path := request.URL.Path

	if strings.HasPrefix(path, "/api/products") {

		if request.Method == "GET" {
			server.handleGET(writer, request)
		} else if request.Method == "POST" {
			server.handlePOST(writer, request)
		} else if request.Method == "PUT" {
			server.handlePUT(writer, request)
		} else if request.Method == "DELETE" {
			server.handleDELETE(writer, request)
		} else {
			notFoundError = getNotFoundResponse()
		}

	} else {
		notFoundError = getNotFoundResponse()
	}

	if notFoundError.responseCode != 0 {
		writeError(writer, notFoundError)
	}
}
