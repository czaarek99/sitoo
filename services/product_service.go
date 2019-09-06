package services

import (
	"errors"
	"fmt"
	"log"
	"sitoo/domain"
	"strings"
)

//TODO: Validate too long and too short strings

type ProductServiceImpl struct {
	Repo domain.ProductRepository
}

func getGenericDatabaseError() error {
	return errors.New("Database error")
}

func handleDatabaseError(err error) {
	log.Println("Database error:")
	log.Println(err.Error())
}

func getAttributeHash(attribute domain.ProductAttribute) string {
	attributeHash := strings.Builder{}
	attributeHash.WriteString(attribute.Name)
	attributeHash.WriteString("_")
	attributeHash.WriteString(attribute.Value)

	return attributeHash.String()
}

func (service ProductServiceImpl) validateBarcodesAreUnique(
	barcodes []string,
) error {

	barcodeSet := map[string]struct{}{}

	for _, barcode := range barcodes {
		barcodeSet[barcode] = struct{}{}
	}

	if len(barcodeSet) < len(barcodes) {
		return errors.New("Barcodes not unique")
	}

	return nil
}

func (service ProductServiceImpl) validateAttributes(
	attributes []domain.ProductAttribute,
) error {

	attributeSet := map[string]struct{}{}

	for _, attribute := range attributes {
		if len(attribute.Name) > 16 {
			return errors.New("Attribute name too long")
		}

		if len(attribute.Value) > 32 {
			return errors.New("Attribute value too long")
		}

		hash := getAttributeHash(attribute)
		attributeSet[hash] = struct{}{}
	}

	if len(attributeSet) < len(attributes) {
		return errors.New("Attributes not unique")
	}

	return nil
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
		handleDatabaseError(err)
		return nil, 0, getGenericDatabaseError()
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
		handleDatabaseError(err)
		return nil, getGenericDatabaseError()
	} else if !exists {
		newErr := errors.New(fmt.Sprintf("Can't find product %v", id))
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
		handleDatabaseError(err)
		return 0, getGenericDatabaseError()
	}

	if productSku != nil {
		return 0, errors.New(fmt.Sprintf("SKU '%s' already exists", product.Sku))
	}

	if len(product.Attributes) > 0 {
		err := service.validateAttributes(product.Attributes)

		if err != nil {
			return 0, err
		}
	}

	if len(product.Barcodes) > 0 {
		err := service.validateBarcodesAreUnique(product.Barcodes)

		if err != nil {
			return 0, err
		}

		barcodes, err := service.Repo.GetBarcodes(product.Barcodes)

		if err != nil {
			return 0, err
		}

		if len(barcodes) > 0 {
			return 0, errors.New("Barcodes not unique")
		}
	}

	id, err := service.Repo.AddProduct(product)

	if err != nil {
		handleDatabaseError(err)
		return 0, getGenericDatabaseError()
	}

	return id, nil
}

func (service ProductServiceImpl) UpdateProduct(
	id domain.ProductId,
	product domain.ProductUpdateInput,
) error {

	exists, err := service.Repo.ProductExists(id)

	if err != nil {
		handleDatabaseError(err)
		return getGenericDatabaseError()
	}

	if !exists {
		return errors.New(fmt.Sprintf("Can't find product %v", id))
	}

	if product.Sku != nil {
		productSku, err := service.Repo.GetSku(*product.Sku)

		if err != nil {
			return err
		}

		if productSku != nil && productSku.ProductID != id {
			return errors.New(fmt.Sprintf("SKU '%s' already exists", product.Sku))
		}
	}

	if len(product.Attributes) > 0 {
		err := service.validateAttributes(product.Attributes)

		if err != nil {
			return err
		}
	}

	if len(product.Barcodes) > 0 {
		err := service.validateBarcodesAreUnique(product.Barcodes)

		if err != nil {
			return err
		}

		barcodes, err := service.Repo.GetBarcodes(product.Barcodes)

		if err != nil {
			return err
		}

		for _, barcode := range barcodes {
			if barcode.ProductID != id {
				return errors.New("Barcodes not unique")
			}
		}
	}

	log.Println("Updating product")

	err = service.Repo.UpdateProduct(id, product)

	if err != nil {
		handleDatabaseError(err)
		return getGenericDatabaseError()
	}

	return nil
}

func (service ProductServiceImpl) DeleteProduct(
	id domain.ProductId,
) error {

	exists, err := service.Repo.ProductExists(id)

	if err != nil {
		handleDatabaseError(err)
		return getGenericDatabaseError()
	}

	if !exists {
		return errors.New(fmt.Sprintf("Product with productId (%v) does not exist", id))
	}

	err = service.Repo.DeleteProduct(id)

	if err != nil {
		handleDatabaseError(err)
		return getGenericDatabaseError()
	}

	return nil
}
