package domain

type ProductId = uint32

type ProductAttribute struct {
	Name  string
	Value string
}

type Product struct {
	ProductID   ProductId
	Title       string
	Sku         string
	Barcodes    []string
	Description string
	Price       string
	Created     uint32
	LastUpdated uint32
	Attributes  []ProductAttribute
}

type ProductAddInput struct {
	Title       string
	Sku         string
	Barcodes    []string
	Description string
	Price       string
	Attributes  []ProductAttribute
}

type ProductUpdateInput struct {
	Title       *string
	Sku         *string
	Barcodes    *[]string
	Description *string
	Price       *string
	Attributes  *[]ProductAttribute
}

type ProductService interface {
	GetProducts(start uint64, num uint64, sku string, barcode string, fields []string) ([]Product, error)
	GetProduct(id ProductId, fields []string) (Product, error)
	AddProduct(product ProductAddInput) (ProductId, error)
	UpdateProduct(id ProductId, product ProductUpdateInput) error
	DeleteProduct(id ProductId) error
}

type ProductRepository interface {
	GetProducts(start uint64, num uint64, sku string, barcode string) ([]Product, error)
	GetProduct(id ProductId) (Product, error)
	AddProduct(product ProductAddInput) (ProductId, error)
	UpdateProduct(id ProductId, product ProductUpdateInput) error
	DeleteProduct(id ProductId) error
	SkuExists(sku string) (bool, error)
}
