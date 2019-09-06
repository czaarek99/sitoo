package services

import (
	"fmt"
	"log"
	"sitoo/domain"
	"sitoo/validation"
)

//TODO: Validate too long and too short strings

type ProductServiceImpl struct {
	Repo domain.ProductRepository
}

func (service ProductServiceImpl) GetProducts(
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

	products, count, err := service.Repo.GetProducts(start, num, sku, barcode, fields)

	if err != nil {
		validation.HandleDatabaseError(err)
		return nil, 0, validation.GetGenericDatabaseError()
	}

	return products, count, nil
}

func (service ProductServiceImpl) GetProduct(
	id domain.ProductId,
	fields []string,
) (*domain.Product, error) {

	log.Println("Requesting single product")

	product, exists, err := service.Repo.GetProduct(id, fields)

	if err != nil {
		validation.HandleDatabaseError(err)
		return nil, validation.GetGenericDatabaseError()
	} else if !exists {
		newErr := fmt.Errorf("Can't find product %v", id)
		return nil, newErr
	}

	return product, nil
}

func (service ProductServiceImpl) AddProduct(
	product domain.ProductAddInput,
) (domain.ProductId, error) {

	log.Println("Adding product")

	productSku, err := service.Repo.GetSku(product.Sku)

	if err != nil {
		validation.HandleDatabaseError(err)
		return 0, validation.GetGenericDatabaseError()
	}

	if productSku != nil {
		return 0, validation.GetSkuAlreadyExistsError(product.Sku)
	}

	if len(product.Attributes) > 0 {
		err := validation.ValidateAttributes(product.Attributes)

		if err != nil {
			return 0, err
		}
	}

	if len(product.Barcodes) > 0 {
		err := validation.ValidateBarcodes(product.Barcodes)

		if err != nil {
			return 0, err
		}

		barcodes, err := service.Repo.GetBarcodes(product.Barcodes)

		if err != nil {
			return 0, err
		}

		if len(barcodes) > 0 {
			return 0, validation.GetBarcodesNotUniqueError()
		}
	}

	id, err := service.Repo.AddProduct(product)

	if err != nil {
		validation.HandleDatabaseError(err)
		return 0, validation.GetGenericDatabaseError()
	}

	return id, nil
}

func (service ProductServiceImpl) UpdateProduct(
	id domain.ProductId,
	product domain.ProductUpdateInput,
) error {

	exists, err := service.Repo.ProductExists(id)

	if err != nil {
		validation.HandleDatabaseError(err)
		return validation.GetGenericDatabaseError()
	}

	if !exists {
		return fmt.Errorf("Can't find product %v", id)
	}

	if product.Sku != nil {
		productSku, err := service.Repo.GetSku(*product.Sku)

		if err != nil {
			return err
		}

		if productSku != nil && productSku.ProductID != id {
			return validation.GetSkuAlreadyExistsError(*product.Sku)
		}
	}

	if len(product.Attributes) > 0 {
		err := validation.ValidateAttributes(product.Attributes)

		if err != nil {
			return err
		}
	}

	if len(product.Barcodes) > 0 {
		err := validation.ValidateBarcodes(product.Barcodes)

		if err != nil {
			return err
		}

		barcodes, err := service.Repo.GetBarcodes(product.Barcodes)

		if err != nil {
			return err
		}

		for _, barcode := range barcodes {
			if barcode.ProductID != id {
				return validation.GetBarcodesNotUniqueError()
			}
		}
	}

	log.Println("Updating product")

	err = service.Repo.UpdateProduct(id, product)

	if err != nil {
		validation.HandleDatabaseError(err)
		return validation.GetGenericDatabaseError()
	}

	return nil
}

func (service ProductServiceImpl) DeleteProduct(
	id domain.ProductId,
) error {

	exists, err := service.Repo.ProductExists(id)

	if err != nil {
		validation.HandleDatabaseError(err)
		return validation.GetGenericDatabaseError()
	}

	if !exists {
		return fmt.Errorf("Product with productId (%v) does not exist", id)
	}

	err = service.Repo.DeleteProduct(id)

	if err != nil {
		validation.HandleDatabaseError(err)
		return validation.GetGenericDatabaseError()
	}

	return nil
}
