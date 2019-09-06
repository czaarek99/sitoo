package validation

import (
	"errors"
	"fmt"
	"log"
	"sitoo/domain"
	"strings"
)

func getAttributeHash(attribute domain.ProductAttribute) string {
	attributeHash := strings.Builder{}
	attributeHash.WriteString(attribute.Name)
	attributeHash.WriteString("_")
	attributeHash.WriteString(attribute.Value)

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

func HandleDatabaseError(err error) {
	log.Println("Database error:")
	log.Println(err.Error())
}

func ValidateBarcodes(barcodes []string) error {

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

	return nil
}

func ValidateAttributes(attributes []domain.ProductAttribute) error {

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

	return nil
}
