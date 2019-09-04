package domain

type Product struct {
	ProductId   int
	Title       string
	Sku         string
	Barcodes    []int
	Description string
	Price       string
	Created     int
	LastUpdated int
}

type ProductService interface {
	GetProducts(start int, num int, sku string, barcode string, fields []string) ([]Product, error)
	GetProduct(id int, fields []string) (Product, error)
	AddProduct(product Product) (int, error)
	UpdateProduct(id int, product Product) error
	DeleteProduct(id int) error
}

type ProductRepository interface {
	GetProducts(start int, num int, sku string, barcode string) ([]Product, error)
	GetProduct(id int) (Product, error)
	AddProduct(product Product) (int, error)
	UpdateProduct(id int, product Product) error
	DeleteProduct(id int) error
	SkuExists(sku string) (bool, error)
}
