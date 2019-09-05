package domain

import "net/http"

type ProductId = uint32

type ProductAttribute struct {
	Name  string
	Value string
}

type Product struct {
	ProductID   ProductId `json: "productId"`
	Title       string    `json: "title"`
	Sku         string    `json: "sku"`
	Barcodes    []string
	Description *string `json: "description"`
	Price       string  `json: "price"`
	Created     uint32  `json: "created`
	LastUpdated *uint32 `json: "lastUpdated"`
	Attributes  []ProductAttribute
}

type ProductAddInput struct {
	Title       string
	Sku         string
	Barcodes    []string
	Description *string
	Price       string
	Attributes  []ProductAttribute
}

type ProductUpdateInput struct {
	Title       *string
	Sku         *string
	Barcodes    []string
	Description *string
	Price       *string
	Attributes  []ProductAttribute
}

type ProductService interface {
	GetProducts(
		start uint64,
		num uint64,
		sku string,
		barcode string,
		fields []string,
	) ([]Product, uint32, error)

	GetProduct(id ProductId, fields []string) (Product, error)
	AddProduct(product ProductAddInput) (ProductId, error)
	UpdateProduct(id ProductId, product ProductUpdateInput) error
	DeleteProduct(id ProductId) error
}

type ProductRepository interface {
	GetProducts(
		start uint64,
		num uint64,
		sku string,
		barcode string,
		fields []string,
	) ([]Product, uint32, error)

	GetProduct(id ProductId, fields []string) (Product, bool, error)
	AddProduct(product ProductAddInput) (ProductId, error)
	UpdateProduct(id ProductId, product ProductUpdateInput) error
	DeleteProduct(id ProductId) error
	SkuExists(sku string) (bool, error)
	ProductExists(id ProductId) (bool, error)
}

type ErrorResponse struct {
	ErrorText    string `json:"errorText"`
	ResponseCode int
}

type ProductServer interface {
	HandleRequest(writer http.ResponseWriter, request *http.Request)
	/*HandleGET(request *http.Request) (interface{}, ErrorResponse)
	HandlePOST(request *http.Request) (uint32, ErrorResponse)
	HandlePUT(request *http.Request) (bool, ErrorResponse)
	HandleDELETE(request *http.Request) (bool, ErrorResponse)*/
}
