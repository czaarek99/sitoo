package repositories

import (
	"api/domain"
	"api/util"
	"database/sql"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type ProductRepositoryImpl struct {
	DB *sql.DB
}

func fieldsToMap(fields []string) map[string]struct{} {

	resultMap := map[string]struct{}{}

	for _, field := range fields {
		resultMap[field] = struct{}{}
	}

	return resultMap
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

func convertSQLDateToTimestamp(date string) (int64, error) {
	createdTime, err := time.Parse("2006-01-02 15:04:05", date)

	if err != nil {
		return 0, err
	}

	return createdTime.Unix(), nil
}

func addToScan(
	toScan []interface{},
	fieldMap map[string]struct{},
	field string,
	address interface{},
) []interface{} {

	if len(fieldMap) == 0 {
		return append(toScan, address)
	}

	_, exists := fieldMap[field]

	if exists {
		return append(toScan, address)
	}

	return toScan
}

func addToSelect(
	toSelect []string,
	fieldMap map[string]struct{},
	field string,
	dbFieldToSelect string,
) []string {

	if len(fieldMap) == 0 {
		return append(toSelect, dbFieldToSelect)
	}

	_, exists := fieldMap[field]

	if exists {
		return append(toSelect, dbFieldToSelect)
	}

	return toSelect
}

func rowToProduct(
	rows *sql.Rows,
	fieldMap map[string]struct{},
) (*domain.Product, error) {
	product := domain.Product{}

	rows.Scan()
	var created string
	var lastUpdated *string

	toScan := []interface{}{}

	toScan = addToScan(toScan, fieldMap, "productId", &product.ProductID)
	toScan = addToScan(toScan, fieldMap, "title", &product.Title)
	toScan = addToScan(toScan, fieldMap, "sku", &product.Sku)
	toScan = addToScan(toScan, fieldMap, "description", &product.Description)
	toScan = addToScan(toScan, fieldMap, "price", &product.Price)
	toScan = addToScan(toScan, fieldMap, "created", &created)
	toScan = addToScan(toScan, fieldMap, "lastUpdated", &lastUpdated)

	err := rows.Scan(toScan...)

	if err != nil {
		return nil, err
	}

	if lastUpdated != nil {
		lastUpdatedTimestamp, err := convertSQLDateToTimestamp(*lastUpdated)

		if err != nil {
			return nil, err
		}

		product.LastUpdated = &lastUpdatedTimestamp
	}

	if created != "" {
		product.Created, err = convertSQLDateToTimestamp(created)
	}

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func getWhereIn(
	column string,
	valueCount int,
) string {

	whereBuilder := strings.Builder{}
	whereBuilder.WriteString("barcode IN(")

	prefix := ""
	for i := 0; i < valueCount; i++ {
		whereBuilder.WriteString(prefix)
		prefix = ","
		whereBuilder.WriteString("?")
	}

	whereBuilder.WriteString(")")

	return whereBuilder.String()
}

func (repo ProductRepositoryImpl) GetProducts(
	start uint64,
	num uint64,
	sku string,
	barcode string,
	fields []string,
) ([]domain.Product, uint32, error) {

	fieldMap := fieldsToMap(fields)

	countQuery := sq.Select("count(*)").
		From("product").
		LeftJoin("product_barcode USING (product_id)")

	toSelect := []string{}

	toSelect = addToSelect(toSelect, fieldMap, "productId", "product.product_id")
	toSelect = addToSelect(toSelect, fieldMap, "title", "product.title")
	toSelect = addToSelect(toSelect, fieldMap, "sku", "product.sku")
	toSelect = addToSelect(toSelect, fieldMap, "description", "product.description")
	toSelect = addToSelect(toSelect, fieldMap, "price", "product.price")
	toSelect = addToSelect(toSelect, fieldMap, "created", "product.created")
	toSelect = addToSelect(toSelect, fieldMap, "lastUpdated", "product.last_updated")

	query := sq.Select(toSelect...).
		LeftJoin("product_barcode USING (product_id)").
		From("product").
		Limit(num).
		Offset(start)

	if sku != "" {
		predicate := sq.Eq{
			"product.sku": sku,
		}

		query = query.Where(predicate)
		countQuery = countQuery.Where(predicate)
	}

	if barcode != "" {
		predicate := sq.Eq{
			"product_barcode.barcode": barcode,
		}

		query = query.Where(predicate)
		countQuery = countQuery.Where(predicate)
	}

	rows, err := query.RunWith(repo.DB).Query()

	defer rows.Close()

	if err != nil {
		return nil, 0, err
	}

	productsMap := map[domain.ProductId]domain.Product{}

	prefix := ""
	inBuilder := strings.Builder{}

	inBuilder.WriteString("product_id IN(")

	for rows.Next() {
		product, err := rowToProduct(rows, fieldMap)

		if err != nil {
			return nil, 0, err
		}

		idString := strconv.FormatUint(uint64(product.ProductID), 10)

		inBuilder.WriteString(prefix)
		prefix = ","
		inBuilder.WriteString(idString)

		productsMap[product.ProductID] = *product
	}

	inBuilder.WriteString(")")

	_, hasBarcodesField := fieldMap["barcodes"]

	if hasBarcodesField {
		barcodeRows, err := sq.Select("product_id", "barcode").
			From("product_barcode").
			Where(inBuilder.String()).
			RunWith(repo.DB).
			Query()

		defer barcodeRows.Close()

		if err != nil {
			return nil, 0, err
		}

		for barcodeRows.Next() {
			var productID uint32
			var barcode string

			err := barcodeRows.Scan(&productID, &barcode)

			if err != nil {
				return nil, 0, err
			}

			product := productsMap[productID]
			product.Barcodes = append(product.Barcodes, barcode)

			productsMap[productID] = product
		}
	}

	_, hasAttributesField := fieldMap["attributes"]

	if hasAttributesField {
		attributeRows, err := sq.Select("product_id", "name", "value").
			From("product_attribute").
			Where(inBuilder.String()).
			RunWith(repo.DB).
			Query()

		defer attributeRows.Close()

		if err != nil {
			return nil, 0, err
		}

		for attributeRows.Next() {
			var productID uint32
			var name string
			var value string

			err := attributeRows.Scan(&productID, &name, &value)

			if err != nil {
				return nil, 0, err
			}

			attribute := domain.ProductAttribute{
				Name:  name,
				Value: value,
			}

			product := productsMap[productID]
			product.Attributes = append(product.Attributes, attribute)

			productsMap[productID] = product
		}
	}

	countQueryString, args, err := countQuery.ToSql()

	if err != nil {
		return nil, 0, err
	}

	count, err := repo.count(countQueryString, args...)

	if err != nil {
		return nil, 0, err
	}

	products := []domain.Product{}

	for _, product := range productsMap {
		products = append(products, product)
	}

	return products, count, nil
}

func (repo ProductRepositoryImpl) GetProduct(
	id domain.ProductId,
	fields []string,
) (*domain.Product, bool, error) {

	fieldMap := fieldsToMap(fields)

	predicate := sq.Eq{
		"product_id": id,
	}

	toSelect := []string{}

	toSelect = addToSelect(toSelect, fieldMap, "productId", "product_id")
	toSelect = addToSelect(toSelect, fieldMap, "title", "title")
	toSelect = addToSelect(toSelect, fieldMap, "sku", "sku")
	toSelect = addToSelect(toSelect, fieldMap, "description", "description")
	toSelect = addToSelect(toSelect, fieldMap, "price", "price")
	toSelect = addToSelect(toSelect, fieldMap, "created", "created")
	toSelect = addToSelect(toSelect, fieldMap, "lastUpdated", "last_updated")

	rows, err := sq.Select(toSelect...).
		From("product").
		Where(predicate).
		RunWith(repo.DB).
		Query()

	defer rows.Close()

	if err != nil {
		return nil, false, err
	}

	exists := rows.Next()

	if !exists {
		return nil, false, nil
	}

	product, err := rowToProduct(rows, fieldMap)

	if err != nil {
		return nil, false, err
	}

	_, hasBarcodesField := fieldMap["barcodes"]

	if hasBarcodesField {
		barcodeRows, err := sq.Select("barcode").
			From("product_barcode").
			Where(predicate).
			RunWith(repo.DB).
			Query()

		defer barcodeRows.Close()

		if err != nil {
			return nil, false, err
		}

		for barcodeRows.Next() {
			var barcode string

			err := barcodeRows.Scan(&barcode)

			if err != nil {
				return nil, false, err
			}

			product.Barcodes = append(product.Barcodes, barcode)
		}
	}

	_, hasAttributeFields := fieldMap["attributes"]

	if hasAttributeFields {
		attributeRows, err := sq.Select("name", "value").
			From("product_attribute").
			Where(predicate).
			RunWith(repo.DB).
			Query()

		defer attributeRows.Close()

		if err != nil {
			return nil, false, err
		}

		for attributeRows.Next() {
			var name string
			var value string

			err := attributeRows.Scan(&name, &value)

			if err != nil {
				return nil, false, err
			}

			product.Attributes = append(product.Attributes, domain.ProductAttribute{
				Name:  name,
				Value: value,
			})
		}
	}

	return product, true, nil
}

func (repo ProductRepositoryImpl) AddProduct(
	product domain.ProductAddInput,
) (domain.ProductId, error) {

	price, err := strconv.ParseFloat(*product.Price, 32)

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

	if product.Barcodes != nil {
		_, err = sq.Delete("product_barcode").Where(predicate).RunWith(tx).Query()

		if err != nil {
			tx.Rollback()
			return err
		}

		if len(product.Barcodes) > 0 {
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
	}

	if product.Attributes != nil {
		_, err = sq.Delete("product_attribute").Where(predicate).RunWith(repo.DB).Query()

		if err != nil {
			tx.Rollback()
			return err
		}

		if len(product.Attributes) > 0 {
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

	tx.Commit()

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

func (repo ProductRepositoryImpl) GetSku(
	sku string,
) (*domain.ProductSku, error) {

	predicate := sq.Eq{
		"sku": sku,
	}

	rows, err := sq.Select("product_id", "sku").
		From("product").
		Where(predicate).
		RunWith(repo.DB).
		Query()

	if err != nil {
		return nil, err
	}

	exists := rows.Next()

	if !exists {
		return nil, nil
	}

	productSku := domain.ProductSku{}

	err = rows.Scan(&productSku.ProductID, &productSku.Sku)

	if err != nil {
		return nil, err
	}

	return &productSku, nil
}

func (repo ProductRepositoryImpl) GetBarcodes(
	barcodes []string,
) ([]domain.ProductBarcode, error) {

	productBarcodes := []domain.ProductBarcode{}

	query := sq.Select("product_id", "barcode").From("product_barcode")

	interfaces := util.StringsToInterfaces(barcodes)
	whereIn := getWhereIn("barcode", len(barcodes))

	query = query.Where(whereIn, interfaces...)

	rows, err := query.RunWith(repo.DB).Query()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		barcode := domain.ProductBarcode{}

		err := rows.Scan(&barcode.ProductID, &barcode.Barcode)

		if err != nil {
			return nil, err
		}

		productBarcodes = append(productBarcodes, barcode)
	}

	return productBarcodes, nil
}
