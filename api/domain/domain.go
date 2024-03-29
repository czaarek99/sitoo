package domain

import (
	"net/http"
)

/*
The domain file defines our Service and Repository
interfaces. It also defines all the interfaces
that are needed to make them work. The domain
is essentially the mold that the project
needs to adhere to and simplifies future updates.
*/

type ProductId = uint32

type ProductAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ProductBarcode struct {
	ProductID ProductId
	Barcode   string
}

type ProductSku struct {
	ProductID ProductId
	Sku       string
}

type Product struct {
	ProductID   ProductId          `json:"productId,omitempty"`
	Title       string             `json:"title,omitempty"`
	Sku         string             `json:"sku,omitempty"`
	Barcodes    []string           `json:"barcodes,omitempty"`
	Description *string            `json:"description,omitempty"`
	Price       string             `json:"price,omitempty"`
	Created     int64              `json:"created,omitempty"`
	LastUpdated *int64             `json:"lastUpdated,omitempty"`
	Attributes  []ProductAttribute `json:"attributes,omitempty"`
}

type ProductAddInput struct {
	Title       string             `json:"title"`
	Sku         string             `json:"sku"`
	Barcodes    []string           `json:"barcodes"`
	Description *string            `json:"description"`
	Price       *string            `json:"price"`
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
	GetBarcodes(barcodes []string) ([]ProductBarcode, error)
	GetSku(sku string) (*ProductSku, error)
	ProductExists(id ProductId) (bool, error)
}

type ProductServer interface {
	HandleRequest(writer http.ResponseWriter, request *http.Request)
}
