package domain

type ProductId = uint32

type Product struct {
	ProductID   ProductId
	Title       string
	Sku         string
	Barcodes    []int
	Description string
	Price       string
	Created     uint32
	LastUpdated uint32
}

type ProductService interface {
	GetProducts(start uint64, num uint64, sku string, barcode string, fields []string) ([]Product, error)
	GetProduct(id ProductId, fields []string) (Product, error)
	AddProduct(product Product) (int, error)
	UpdateProduct(id ProductId, product Product) error
	DeleteProduct(id ProductId) error
}

type ProductRepository interface {
	GetProducts(start uint64, num uint64, sku string, barcode string) ([]Product, error)
	GetProduct(id ProductId) (Product, error)
	AddProduct(product Product) (ProductId, error)
	UpdateProduct(id ProductId, product Product) error
	DeleteProduct(id ProductId) error
	SkuExists(sku string) (bool, error)
}
