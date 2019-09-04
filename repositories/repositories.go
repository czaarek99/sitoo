package repositories

//When this is ready make sure to rebase this commit

import (
	"database/sql"
	"sitoo/domain"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

type Repository struct {
	db *sql.DB
}

type DatabaseRow struct {
	productID      domain.ProductId
	title          string
	sku            string
	description    string
	price          string
	created        uint32
	lastUpdated    uint32
	barcode        string
	attributeName  string
	attributeValue string
}

type DatabaseProduct struct {
	productID   domain.ProductId
	title       string
	sku         string
	description string
	price       string
	created     uint32
	lastUpdated uint32

	barcodes   map[string]struct{}
	attributes map[string]domain.ProductAttribute
}

func getAttributeHash(attribute domain.ProductAttribute) string {
	attributeHash := strings.Builder{}
	attributeHash.WriteString(attribute.Name)
	attributeHash.WriteString("_")
	attributeHash.WriteString(attribute.Value)

	return attributeHash.String()
}

/*
Could be optimized to use 3 queries instead of one.
That adds a lot of complexity to the problem so we'll
skip that here.
*/
func (repo Repository) GetProducts(
	start uint64,
	num uint64,
	sku string,
	barcode string,
) ([]domain.Product, error) {

	query := sq.Select(
		"product.product_id",
		"product.title",
		"product.sku",
		"product.description",
		"product.price",
		"product.created",
		"product.last_updated",
		"product_barcode.barcode",
		"product_attribute.name",
		"product_attribute.value",
	).
		LeftJoin("product_barcode USING (product_id)").
		LeftJoin("product_attribute USING (product_id)").
		From("product").
		Limit(num).
		Offset(start)

	if sku != "" {
		query = query.Where(sq.Eq{
			sku: sku,
		})
	}

	if barcode != "" {
		query = query.Where(sq.Eq{
			barcode: barcode,
		})
	}

	rows, err := query.RunWith(repo.db).Query()

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	productMap := make(map[domain.ProductId]DatabaseProduct)

	for rows.Next() {
		row := DatabaseRow{}

		rows.Scan(
			&row.productID,
			&row.title,
			&row.sku,
			&row.description,
			&row.price,
			&row.created,
			&row.lastUpdated,
			&row.barcode,
			&row.attributeName,
			&row.attributeValue,
		)

		_, exists := productMap[row.productID]

		if !exists {
			barcodes := make(map[string]struct{})
			attributes := make(map[string]domain.ProductAttribute)

			productMap[row.productID] = DatabaseProduct{
				productID:   row.productID,
				title:       row.title,
				sku:         row.sku,
				description: row.description,
				price:       row.price,
				created:     row.created,
				lastUpdated: row.lastUpdated,
				barcodes:    barcodes,
				attributes:  attributes,
			}
		}

		mapEntry, _ := productMap[row.productID]

		mapEntry.barcodes[row.barcode] = struct{}{}

		attribute := domain.ProductAttribute{
			Name:  row.attributeName,
			Value: row.attributeValue,
		}

		hash := getAttributeHash(attribute)

		mapEntry.attributes[hash] = attribute
	}

	results := make([]domain.Product, 5)

	for _, value := range productMap {
		barcodeSlice := make([]string, 5)
		attributeSlice := make([]domain.ProductAttribute, 5)

		for key, _ := range value.barcodes {
			barcodeSlice = append(barcodeSlice, key)
		}

		for _, attribute := range value.attributes {
			attributeSlice = append(attributeSlice, attribute)
		}

		results = append(results, domain.Product{
			ProductID:   value.productID,
			Title:       value.title,
			Sku:         value.sku,
			Description: value.description,
			Price:       value.price,
			Created:     value.created,
			LastUpdated: value.lastUpdated,
			Barcodes:    barcodeSlice,
			Attributes:  attributeSlice,
		})
	}

	return results, nil
}
