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
	GetProducts(start int, num int, sku int, barcode int, fields []string) ([]Product, error)
	GetProduct(id int, fields []string) (Product, error)
	AddProduct(product Product) (int, error)
	UpdateProduct(product Product) error
	DeleteProduct() error
}

type ProductRepository interface {
	GetProducts(start int, num int, sku int, barcode int) ([]Product, error)
	GetProduct(id int) (Product, error)
	AddProduct(product Product) (int, error)
	UpdateProduct(product Product) error
	DeleteProduct() error
}
