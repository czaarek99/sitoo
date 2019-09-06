package domain

import (
	"net/http"
)

type ProductId = uint32

type ProductAttribute struct {
	Name  string
	Value string
}

type Product struct {
	ProductID   ProductId          `json:"productId"`
	Title       string             `json:"title"`
	Sku         string             `json:"sku"`
	Barcodes    []string           `json:"barcodes"`
	Description *string            `json:"description"`
	Price       string             `json:"price"`
	Created     int64              `json:"created"`
	LastUpdated *int64             `json:"lastUpdated"`
	Attributes  []ProductAttribute `json:"attributes"`
}

type ProductAddInput struct {
	Title       string             `json:"title"`
	Sku         string             `json:"sku"`
	Barcodes    []string           `json:"barcodes"`
	Description *string            `json:"description"`
	Price       string             `json:"price"`
	Attributes  []ProductAttribute `json:"attributes"`
}

type ProductUpdateInput struct {
	Title       *string            `json:"title"`
	Sku         *string            `json:"sku"`
	Barcodes    []string           `json:"barcodes"`
	Description *string            `json:"description"`
	Price       *string            `json:"price"`
	Attributes  []ProductAttribute `json:"attributes"`
}

type ProductService interface {
	GetProducts(
		start uint64,
		num uint64,
		sku string,
		barcode string,
		fields []string,
	) ([]Product, uint32, error)

	GetProduct(id ProductId, fields []string) (*Product, error)
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

	GetProduct(id ProductId, fields []string) (*Product, bool, error)
	AddProduct(product ProductAddInput) (ProductId, error)
	UpdateProduct(id ProductId, product ProductUpdateInput) error
	DeleteProduct(id ProductId) error
	SkuExists(sku string) (bool, error)
	BarcodesExist(barcode []string) (bool, error)
	ProductExists(id ProductId) (bool, error)
	AttributesExist(id ProductId, attributes []ProductAttribute) (bool, error)
}

type ErrorResponse struct {
	ErrorText    string `json:"errorText"`
	ResponseCode int    `json:"-"`
}

type ProductServer interface {
	HandleRequest(writer http.ResponseWriter, request *http.Request)
}
