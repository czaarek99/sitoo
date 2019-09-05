package servers

import (
	"net/http"
	"sitoo/domain"
)

type Server struct {
	service domain.ProductService
}

func (server Server) HandleGET(
	request *http.Request,
) (interface{}, domain.ErrorResponse) {

	return nil, domain.ErrorResponse{}

}

func (server Server) HandlePOST(
	request *http.Request,
) (uint32, domain.ErrorResponse) {

	return 0, domain.ErrorResponse{}
}

func (server Server) HandlePUT(
	request *http.Request,
) (bool, domain.ErrorResponse) {

	return false, domain.ErrorResponse{}
}

func (server Server) HandleDELETE(
	request *http.Request,
) (bool, domain.ErrorResponse) {

	return false, domain.ErrorResponse{}
}
