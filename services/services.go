package services

import (
	"errors"
	"fmt"
	"log"
	"sitoo/domain"
)

type Service struct {
	repo domain.ProductRepository
}

//TODO: Implement all validation

func (service Service) GetProducts(
	start int,
	num int,
	sku int,
	barcode int,
	fields []string,
) ([]domain.Product, error) {

	log.Println("Requesting multiple products")

	return service.repo.GetProducts(start, num, sku, barcode)
}

func (service Service) GetProduct(
	id int,
	fields []string,
) (domain.Product, error) {

	log.Println("Requesting single product")

	product, err := service.repo.GetProduct(id)

	if err != nil {
		return product, err
	} else {
		newErr := errors.New(fmt.Sprintf("Can't find product %v", id))

		return product, newErr
	}

}

func (service Service) AddProduct(
	product domain.Product,
) (int, error) {

	log.Println("Adding product")

	exists, err := service.repo.SkuExists(product.Sku)

	if err != nil {
		return -1, err
	}

	if exists {
		return -1, errors.New(fmt.Sprintf("SKU '%s' already exists", product.Sku))
	}

	return service.repo.AddProduct(product)
}

func (service Service) UpdateProduct(
	id int,
	product domain.Product,
) error {

	log.Println("Updating product")

	return service.repo.UpdateProduct(id, product)
}

func (service Service) DeleteProduct(
	id int,
) error {

	return service.repo.DeleteProduct(id)

}
