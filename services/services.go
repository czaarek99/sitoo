package services

import (
	"errors"
	"fmt"
	"log"
	"sitoo/domain"
)

type Service struct {
	Repo domain.ProductRepository
}

//TODO: Implement all validation
//Catch database errors and print to console instead

func (service Service) GetProducts(
	start uint64,
	num uint64,
	sku string,
	barcode string,
	fields []string,
) ([]domain.Product, uint32, error) {

	log.Println("Requesting multiple products")

	if num == 0 {
		num = 10
	}

	return service.Repo.GetProducts(start, num, sku, barcode, fields)
}

func (service Service) GetProduct(
	id domain.ProductId,
	fields []string,
) (domain.Product, error) {

	log.Println("Requesting single product")

	product, exists, err := service.Repo.GetProduct(id, fields)

	if err != nil {
		return product, err
	} else if !exists {
		newErr := errors.New(fmt.Sprintf("Can't find product %v", id))
		return product, newErr
	}

	return product, nil
}

func (service Service) AddProduct(
	product domain.ProductAddInput,
) (domain.ProductId, error) {

	log.Println("Adding product")

	exists, err := service.Repo.SkuExists(product.Sku)

	if err != nil {
		return 0, err
	}

	if exists {
		return 0, errors.New(fmt.Sprintf("SKU '%s' already exists", product.Sku))
	}

	return service.Repo.AddProduct(product)
}

func (service Service) UpdateProduct(
	id domain.ProductId,
	product domain.ProductUpdateInput,
) error {

	exists, err := service.Repo.ProductExists(id)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New(fmt.Sprintf("Can't find product %v", id))
	}

	log.Println("Updating product")

	return service.Repo.UpdateProduct(id, product)
}

func (service Service) DeleteProduct(
	id domain.ProductId,
) error {

	exists, err := service.Repo.ProductExists(id)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New(fmt.Sprintf("Product with productId (%v) does not exist", id))
	}

	return service.Repo.DeleteProduct(id)

}
