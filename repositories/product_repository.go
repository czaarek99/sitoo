package repositories

import (
	"database/sql"
	"sitoo/domain"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type ProductRepositoryImpl struct {
	DB *sql.DB
}

//TODO: Handle fields in database
func (repo ProductRepositoryImpl) getTotalCount() (uint32, error) {
	count, err := repo.count("SELECT COUNT(*) as count FROM product")
	return count, err
}

func (repo ProductRepositoryImpl) count(
	query string,
	values ...interface{},
) (uint32, error) {

	rows, err := repo.DB.Query(query, values...)

	defer rows.Close()

	if err != nil {
		return 0, err
	}

	var count uint32

	rows.Next()
	err = rows.Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func getAttributeHash(attribute domain.ProductAttribute) string {
	attributeHash := strings.Builder{}
	attributeHash.WriteString(attribute.Name)
	attributeHash.WriteString("_")
	attributeHash.WriteString(attribute.Value)

	return attributeHash.String()
}

//TODO: Refactor and just make 3 queries instead
func rowsToProducts(rows *sql.Rows) ([]domain.Product, uint32, error) {
	results := make([]domain.Product, 0)

	var rowCount uint32

	first := true
	barcodes := make(map[string]struct{})
	attributes := make(map[string]domain.ProductAttribute)

	index := 0
	var prevId uint32

	for rows.Next() {
		rowCount++

		productEntity := domain.Product{}

		var barcode string
		var attributeName string
		var attributeValue string

		err := rows.Scan(
			&productEntity.ProductID,
			&productEntity.Title,
			&productEntity.Sku,
			productEntity.Description,
			&productEntity.Price,
			&productEntity.Created,
			productEntity.LastUpdated,
			&barcode,
			&attributeName,
			&attributeValue,
		)

		if err != nil {
			return nil, 0, err
		}

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

	return results, rowCount, nil
}

/*
Could be optimized to use 3 queries instead of one.
That adds a lot of complexity to the problem so we'll
skip that here.
*/
func (repo ProductRepositoryImpl) GetProducts(
	start uint64,
	num uint64,
	sku string,
	barcode string,
	fields []string,
) ([]domain.Product, uint32, error) {

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

	rows, err := query.RunWith(repo.DB).Query()

	defer rows.Close()

	if err != nil {
		return nil, 0, err
	}

	products, _, err := rowsToProducts(rows)

	if err != nil {
		return nil, 0, err
	}

	count, err := repo.getTotalCount()

	if err != nil {
		return nil, 0, err
	}

	return products, count, nil
}

func (repo ProductRepositoryImpl) GetProduct(
	id domain.ProductId,
	fields []string,
) (domain.Product, bool, error) {

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
		RunWith(repo.DB).
		Query()

	defer rows.Close()

	if err != nil {
		return domain.Product{}, false, err
	}

	products, count, err := rowsToProducts(rows)

	if len(products) == 0 {
		return domain.Product{}, false, nil
	}

	return products[0], count > 0, err
}

//TODO: Figure out how to get errors from insert query
func (repo ProductRepositoryImpl) AddProduct(
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

	tx, err := repo.DB.Begin()

	if err != nil {
		return 0, err
	}

	query, args, err := sq.Insert("product").
		Columns(
			"title",
			"sku",
			"description",
			"price",
			"created",
		).
		Values(product.Title, product.Sku, description, price, time.Now()).
		ToSql()

	if err != nil {
		return 0, err
	}

	res, err := tx.Exec(query, args...)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	id, _ := res.LastInsertId()

	var productID = domain.ProductId(id)

	if len(product.Barcodes) > 0 {
		barcodeInsert := sq.Insert("product_barcode").Columns("product_id", "barcode")

		for _, barcode := range product.Barcodes {
			barcodeInsert = barcodeInsert.Values(productID, barcode)
		}

		_, err := barcodeInsert.RunWith(tx).Query()

		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	if len(product.Attributes) > 0 {
		attributeInsert := sq.Insert("product_attribute").Columns("product_id", "name", "value")

		for _, attribute := range product.Attributes {
			attributeInsert = attributeInsert.Values(productID, attribute.Name, attribute.Value)
		}

		_, err := attributeInsert.RunWith(tx).Query()

		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	err = tx.Commit()

	if err != nil {
		return 0, err
	}

	return productID, nil
}

//TODO: Explain why we ignore a sku of nil later
func (repo ProductRepositoryImpl) UpdateProduct(
	id domain.ProductId,
	product domain.ProductUpdateInput,
) error {

	predicate := sq.Eq{
		"product_id": id,
	}

	query := sq.Update("product").Set("last_updated", time.Now()).Where(predicate)

	if product.Title != nil {
		query = query.Set("title", product.Title)
	}

	if product.Sku != nil {
		query = query.Set("sku", product.Sku)
	}

	if product.Description != nil {
		query = query.Set("description", product.Description)
	}

	if product.Price != nil {
		query = query.Set("price", product.Price)
	}

	tx, err := repo.DB.Begin()

	if err != nil {
		return err
	}

	_, err = query.RunWith(tx).Query()

	if err != nil {
		tx.Rollback()
		return err
	}

	if len(product.Barcodes) > 0 {
		_, err = sq.Delete("product_barcode").Where(predicate).RunWith(tx).Query()

		if err != nil {
			tx.Rollback()
			return err
		}

		barcodeInsert := sq.Insert("product_barcode").Columns("product_id", "barcode")

		for _, barcode := range product.Barcodes {
			barcodeInsert = barcodeInsert.Values(id, barcode)
		}

		_, err = barcodeInsert.RunWith(tx).Query()

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(product.Attributes) > 0 {
		_, err = sq.Delete("product_attribute").Where(predicate).RunWith(repo.DB).Query()

		if err != nil {
			tx.Rollback()
			return err
		}

		attribtueInsert := sq.Insert("product_attribute").Columns("product_id", "name", "value")

		for _, attribute := range product.Attributes {
			attribtueInsert = attribtueInsert.Values(id, attribute.Name, attribute.Value)
		}

		_, err := attribtueInsert.RunWith(repo.DB).Query()

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (repo ProductRepositoryImpl) DeleteProduct(
	id domain.ProductId,
) error {
	predicate := sq.Eq{
		"product_id": id,
	}

	tx, err := repo.DB.Begin()

	if err != nil {
		return err
	}

	_, err = sq.Delete("product").Where(predicate).RunWith(tx).Query()

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = sq.Delete("product_barcode").Where(predicate).RunWith(tx).Query()

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = sq.Delete("product_attribute").Where(predicate).RunWith(tx).Query()

	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (repo ProductRepositoryImpl) ProductExists(
	id domain.ProductId,
) (bool, error) {

	count, err := repo.count("SELECT COUNT(*) as count FROM product WHERE product_id = ?", id)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo ProductRepositoryImpl) SkuExists(
	sku string,
) (bool, error) {

	count, err := repo.count("SELECT COUNT(*) as count FROM product WHERE sku = ?", sku)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
