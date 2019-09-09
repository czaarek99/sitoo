package validation

import (
	"api/domain"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func getAttributeHash(attribute domain.ProductAttribute) string {
	attributeHash := strings.Builder{}
	attributeHash.WriteString(strings.ToLower(attribute.Name))
	attributeHash.WriteString("_")
	attributeHash.WriteString(strings.ToLower(attribute.Value))

	return attributeHash.String()
}

func GetGenericDatabaseError() error {
	return errors.New("Database error")
}

func GetBarcodesNotUniqueError() error {
	return errors.New("Barcodes not unique")
}

func GetSkuAlreadyExistsError(sku string) error {
	return fmt.Errorf("SKU '%s' already exists", sku)
}

func validateBarcodes(barcodes []string) error {

	if len(barcodes) > 0 {
		barcodeSet := map[string]struct{}{}

		for _, barcode := range barcodes {
			if len(barcode) > 32 {
				return fmt.Errorf("Barcode (%s) is longer than max of 32 characters", barcode)
			}

			barcodeSet[barcode] = struct{}{}
		}

		if len(barcodeSet) < len(barcodes) {
			return GetBarcodesNotUniqueError()
		}
	}

	return nil
}

func validateAttributes(attributes []domain.ProductAttribute) error {

	if len(attributes) > 0 {
		attributeSet := map[string]struct{}{}

		for _, attribute := range attributes {
			if len(attribute.Name) > 16 {
				return fmt.Errorf("Attribute name (%s) is longer than max of 16 characters", attribute.Name)
			}

			if len(attribute.Value) > 32 {
				return fmt.Errorf("Attribute value (%s) is longer than max of 32 characters", attribute.Value)
			}

			hash := getAttributeHash(attribute)
			attributeSet[hash] = struct{}{}
		}

		if len(attributeSet) < len(attributes) {
			return errors.New("Attributes not unique")
		}
	}

	return nil
}

func validateTitle(title string) error {

	if len(title) == 0 {
		return errors.New("Title can not be empty")
	}

	if len(title) > 32 {
		return fmt.Errorf("Product title (%s) is longer than max of 32 characters", title)
	}

	return nil
}

func validateSku(sku string) error {

	if len(sku) == 0 {
		return errors.New("Sku can not be empty")
	}

	if len(sku) > 32 {
		return fmt.Errorf("Product sku (%s) is longer than max of 32 characters", sku)
	}

	return nil
}

func validateDescription(description string) error {
	if len(description) > 1024 {
		return fmt.Errorf("Product description (%s) is longer than max of 32 characters", description)
	}

	return nil
}

func validatePrice(price string) error {

	float, err := strconv.ParseFloat(price, 10)

	if err != nil {
		return fmt.Errorf("Product price (%s) is not a valid decimal", price)
	}

	if float > math.Pow(10, 7) {
		return fmt.Errorf("Product price (%s) is too big", price)
	}

	if float < 0 {
		return fmt.Errorf("Product price (%s) can not be negative", price)
	}

	return nil
}

func ValidateFields(fields []string) error {
	allowedFields := map[string]struct{}{}

	allowedFields["productId"] = struct{}{}
	allowedFields["title"] = struct{}{}
	allowedFields["sku"] = struct{}{}
	allowedFields["barcodes"] = struct{}{}
	allowedFields["description"] = struct{}{}
	allowedFields["attributes"] = struct{}{}
	allowedFields["price"] = struct{}{}
	allowedFields["created"] = struct{}{}
	allowedFields["lastUpdated"] = struct{}{}

	for _, field := range fields {
		_, ok := allowedFields[field]

		if !ok {
			return fmt.Errorf("Unknown field (%s) field list", field)
		}
	}

	return nil
}

func ValidateProductUpdate(changes domain.ProductUpdateInput) error {

	if changes.Title != nil {

		err := validateTitle(*changes.Title)

		if err != nil {
			return err
		}
	}

	if changes.Sku != nil {
		err := validateSku(*changes.Sku)

		if err != nil {
			return err
		}
	}

	if changes.Description != nil {
		err := validateDescription(*changes.Description)

		if err != nil {
			return err
		}
	}

	if changes.Price != nil {
		err := validatePrice(*changes.Price)

		if err != nil {
			return err
		}
	}

	err := validateBarcodes(changes.Barcodes)

	if err != nil {
		return err
	}

	err = validateAttributes(changes.Attributes)

	if err != nil {
		return err
	}

	return nil
}

func ValidateNewProduct(product domain.ProductAddInput) error {

	err := validateTitle(product.Title)

	if err != nil {
		return err
	}

	err = validateSku(product.Sku)

	if err != nil {
		return err
	}

	if product.Description != nil {
		err := validateDescription(*product.Description)

		if err != nil {
			return err
		}
	}

	if product.Price != nil {
		err := validatePrice(*product.Price)

		if err != nil {
			return err
		}
	}

	err = validateBarcodes(product.Barcodes)

	if err != nil {
		return err
	}

	err = validateAttributes(product.Attributes)

	if err != nil {
		return err
	}

	return nil
}
