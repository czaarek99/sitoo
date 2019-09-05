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

func getNotFoundResponse() domain.ErrorResponse {
	return domain.ErrorResponse{
		ErrorText:    "Not found",
		ResponseCode: 404,
	}
}

func writeError(
	writer http.ResponseWriter,
	errorResponse domain.ErrorResponse,
) {
	writeJson(writer, errorResponse, errorResponse.ResponseCode)
}

func writeJson(
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

func getProductIdFromPath(
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

	productID, err := getProductIdFromPath(request.URL.Path)

	if err == nil {
		parsed.productID = domain.ProductId(productID)
		parsed.getType = singleGET
	} else {
		query := request.URL.Query()
		start, err := strconv.ParseUint(query.Get("start"), 10, 64)

		if err != nil {
			parsed.start = start
		}

		num, numErr := strconv.ParseUint(query.Get("num"), 10, 64)

		if numErr != nil {
			parsed.num = num
		}

		delimitedFields := query.Get("fields")

		if delimitedFields != "" {
			parsed.fields = strings.Split(delimitedFields, ",")
		}

		parsed.sku = query.Get("sku")
		parsed.barcode = query.Get("barcode")
		parsed.getType = multipleGET
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
			writeJson(writer, product, http.StatusOK)
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

			writeJson(writer, envelope, http.StatusOK)
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
	writer.Write([]byte(idString))
	writer.WriteHeader(http.StatusOK)
}

func (server Server) handlePUT(
	writer http.ResponseWriter,
	request *http.Request,
) {

	id, err := getProductIdFromPath(request.URL.Path)

	if err != nil {
		writeError(writer, getBadRequestResponse("Bad Request"))
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

	id, err := getProductIdFromPath(request.URL.Path)

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

	var notFoundError domain.ErrorResponse

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

	if notFoundError.ResponseCode != 0 {
		writeError(writer, notFoundError)
	}
}
