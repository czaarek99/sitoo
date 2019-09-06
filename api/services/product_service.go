package services

import (
	"fmt"
	"log"
	"sitoo/domain"
	"sitoo/util"
	"sitoo/validation"
)

type ProductServiceImpl struct {
	Repo     domain.ProductRepository
	Metadata util.Metadata
}

func (service ProductServiceImpl) log(
	format string,
	values ...interface{},
) {

	message := fmt.Sprintf(format, values...)
	log.Printf("requestId=%v data/message=%s", service.Metadata.RequestID, message)
}

func (service ProductServiceImpl) handleDatabaseError(
	err error,
) {
	service.log("Database error %s", err.Error())
}

func (service ProductServiceImpl) GetProducts(
	start uint64,
	num uint64,
	sku string,
	barcode string,
	fields []string,
) ([]domain.Product, uint32, error) {

	service.log("Requesting multiple products")

	err := validation.ValidateFields(fields)

	if err != nil {
		service.log("Validation failed")

		return nil, 0, err
	}

	if num == 0 {
		service.log("Default value init for sum")

		num = 10
	}

	products, count, err := service.Repo.GetProducts(start, num, sku, barcode, fields)

	if err != nil {
		service.handleDatabaseError(err)
		return nil, 0, validation.GetGenericDatabaseError()
	}

	service.log("Sending back products")

	return products, count, nil
}

func (service ProductServiceImpl) GetProduct(
	id domain.ProductId,
	fields []string,
) (*domain.Product, error) {

	service.log("Requested single product with id %v", id)

	err := validation.ValidateFields(fields)

	if err != nil {
		service.log("Validation failed")

		return nil, err
	}

	product, exists, err := service.Repo.GetProduct(id, fields)

	if err != nil {
		service.handleDatabaseError(err)
		return nil, validation.GetGenericDatabaseError()
	} else if !exists {
		service.log("Can't find product with id %v", id)

		newErr := fmt.Errorf("Can't find product %v", id)
		return nil, newErr
	}

	return product, nil
}

func (service ProductServiceImpl) AddProduct(
	product domain.ProductAddInput,
) (domain.ProductId, error) {

	service.log("Adding product")

	err := validation.ValidateNewProduct(product)

	if err != nil {
		service.log("Failed validation")

		return 0, err
	}

	productSku, err := service.Repo.GetSku(product.Sku)

	if err != nil {
		service.handleDatabaseError(err)
		return 0, validation.GetGenericDatabaseError()
	}

	if productSku != nil {
		service.log("Sku (%s) already exists", product.Sku)

		return 0, validation.GetSkuAlreadyExistsError(product.Sku)
	}

	if len(product.Barcodes) > 0 {
		barcodes, err := service.Repo.GetBarcodes(product.Barcodes)

		if err != nil {
			service.handleDatabaseError(err)
			return 0, validation.GetGenericDatabaseError()
		}

		if len(barcodes) > 0 {
			service.log("Barcodes not unique")

			return 0, validation.GetBarcodesNotUniqueError()
		}
	}

	id, err := service.Repo.AddProduct(product)

	if err != nil {
		service.handleDatabaseError(err)
		return 0, validation.GetGenericDatabaseError()
	}

	service.log("Added product")

	return id, nil
}

func (service ProductServiceImpl) UpdateProduct(
	id domain.ProductId,
	product domain.ProductUpdateInput,
) error {

	service.log("Updating product with id (%v)", id)

	exists, err := service.Repo.ProductExists(id)

	if err != nil {
		service.handleDatabaseError(err)
		return validation.GetGenericDatabaseError()
	}

	if !exists {
		service.log("Cant find product")

		return fmt.Errorf("Can't find product %v", id)
	}

	err = validation.ValidateProductUpdate(product)

	if err != nil {
		return err
	}

	if product.Sku != nil {
		productSku, err := service.Repo.GetSku(*product.Sku)

		if err != nil {
			service.handleDatabaseError(err)
			return validation.GetGenericDatabaseError()
		}

		if productSku != nil && productSku.ProductID != id {
			service.log("Sku already exists")
			return validation.GetSkuAlreadyExistsError(*product.Sku)
		}
	}

	if len(product.Barcodes) > 0 {
		barcodes, err := service.Repo.GetBarcodes(product.Barcodes)

		if err != nil {
			service.handleDatabaseError(err)
			return validation.GetGenericDatabaseError()
		}

		for _, barcode := range barcodes {
			if barcode.ProductID != id {
				service.log("Barcodes not unique")

				return validation.GetBarcodesNotUniqueError()
			}
		}
	}

	err = service.Repo.UpdateProduct(id, product)

	if err != nil {
		service.handleDatabaseError(err)
		return validation.GetGenericDatabaseError()
	}

	service.log("Updated product")

	return nil
}

func (service ProductServiceImpl) DeleteProduct(
	id domain.ProductId,
) error {

	service.log("Deleting product with id: (%v)", id)

	exists, err := service.Repo.ProductExists(id)

	if err != nil {
		service.handleDatabaseError(err)
		return validation.GetGenericDatabaseError()
	}

	if !exists {
		service.log("Product does not exist")

		return fmt.Errorf("Product with productId (%v) does not exist", id)
	}

	err = service.Repo.DeleteProduct(id)

	if err != nil {
		service.handleDatabaseError(err)
		return validation.GetGenericDatabaseError()
	}

	service.log("Deleted product")

	return nil
}
