package repositories

import (
	"database/sql"
	"sitoo/domain"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

type Repository struct {
	db *sql.DB
}

func getAttributeHash(attribute domain.ProductAttribute) string {
	attributeHash := strings.Builder{}
	attributeHash.WriteString(attribute.Name)
	attributeHash.WriteString("_")
	attributeHash.WriteString(attribute.Value)

	return attributeHash.String()
}

func rowsToProducts(rows *sql.Rows) []domain.Product {
	defer rows.Close()

	results := make([]domain.Product, 1)

	first := true
	barcodes := make(map[string]struct{})
	attributes := make(map[string]domain.ProductAttribute)

	index := 0
	var prevId uint32

	for rows.Next() {
		productEntity := domain.Product{}

		var barcode string
		var attributeName string
		var attributeValue string

		rows.Scan(
			&productEntity.ProductID,
			&productEntity.Title,
			&productEntity.Sku,
			productEntity.Description,
			&productEntity.Price,
			&productEntity.Created,
			&productEntity.LastUpdated,
			&barcode,
			&attributeName,
			&attributeValue,
		)

		isNewId := prevId != productEntity.ProductID

		if first || isNewId {
			results = append(results, productEntity)
		}

		first = false

		if isNewId {
			barcodeSlice := make([]string, 1)
			attributeSlice := make([]domain.ProductAttribute, 1)

			for key := range barcodes {
				barcodeSlice = append(barcodeSlice, key)
			}

			for _, value := range attributes {
				attributeSlice = append(attributeSlice, value)
			}

			results[index].Barcodes = barcodeSlice
			results[index].Attributes = attributeSlice

			barcodes = make(map[string]struct{})
			attributes = make(map[string]domain.ProductAttribute)

			index++
		}

		attribute := domain.ProductAttribute{
			Name:  attributeName,
			Value: attributeValue,
		}

		attributeHash := getAttributeHash(attribute)
		attributes[attributeHash] = attribute
		barcodes[barcode] = struct{}{}

		prevId = productEntity.ProductID
	}

	return results
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
		Offset(start).
		OrderBy("product.product_id")

	if sku != "" {
		query = query.Where(sq.Eq{
			"sku": sku,
		})
	}

	if barcode != "" {
		query = query.Where(sq.Eq{
			"barcode": barcode,
		})
	}

	rows, err := query.RunWith(repo.db).Query()

	if err != nil {
		return nil, err
	}

	return rowsToProducts(rows), nil
}

func (repo Repository) GetProduct(
	id domain.ProductId,
) (domain.Product, error) {

	rows, err := sq.Select(
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
		Where(sq.Eq{
			"product_id": id,
		}).
		OrderBy("product.product_id").
		RunWith(repo.db).
		Query()

	if err != nil {
		return domain.Product{}, err
	}

	products := rowsToProducts(rows)

	return products[0], nil
}

//TODO: Figure out how to get errors from insert query
func (repo Repository) AddProduct(
	product domain.ProductAddInput,
) (domain.ProductId, error) {

	price, err := strconv.ParseFloat(product.Price, 32)

	if err != nil {
		return 0, err
	}

	var description sql.NullString

	if product.Description != nil {
		description = sql.NullString{
			String: *product.Description,
			Valid:  true,
		}
	}

	insert := sq.Insert("products").
		Columns(
			"title",
			"sku",
			"description",
			"price",
		).
		Values(product.Title, product.Sku, description, price).
		RunWith(repo.db)

	var productID domain.ProductId
	insert.QueryRow().Scan(&productID)

	if len(product.Barcodes) > 0 {
		barcodeInsert := sq.Insert("product_barcode").Columns("product_id", "barcode")

		for _, barcode := range product.Barcodes {
			barcodeInsert = barcodeInsert.Values(productID, barcode)
		}

		barcodeInsert.RunWith(repo.db)
	}

	if len(product.Attributes) > 0 {
		attributeInsert := sq.Insert("product_attribute").Columns("product_id", "name", "value")

		for _, attribute := range product.Attributes {
			attributeInsert = attributeInsert.Values(productID, attribute.Name, attribute.Value)
		}

		attributeInsert.RunWith(repo.db)
	}

	return productID, nil
}

func (repo Repository) UpdateProduct(
	id domain.ProductId,
	product domain.ProductUpdateInput,
) error {

	return nil

}

func (repo Repository) DeleteProduct(
	id domain.ProductId,
) error {

	return nil

}
