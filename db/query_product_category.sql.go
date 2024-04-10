// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: query_product_category.sql

package db

import (
	"context"
	"database/sql"
)

const createProductCategory = `-- name: CreateProductCategory :one
INSERT INTO product_category (
	category,
	subcategory
) VALUES (
	?, ?
) RETURNING id, category, subcategory
`

type CreateProductCategoryParams struct {
	Category    sql.NullString
	Subcategory sql.NullString
}

func (q *Queries) CreateProductCategory(ctx context.Context, arg CreateProductCategoryParams) (ProductCategory, error) {
	row := q.db.QueryRowContext(ctx, createProductCategory, arg.Category, arg.Subcategory)
	var i ProductCategory
	err := row.Scan(&i.ID, &i.Category, &i.Subcategory)
	return i, err
}

const getProductCategories = `-- name: GetProductCategories :many
SELECT id, category, subcategory
FROM product_category
ORDER BY category DESC
`

func (q *Queries) GetProductCategories(ctx context.Context) ([]ProductCategory, error) {
	rows, err := q.db.QueryContext(ctx, getProductCategories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ProductCategory
	for rows.Next() {
		var i ProductCategory
		if err := rows.Scan(&i.ID, &i.Category, &i.Subcategory); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProductCategory = `-- name: GetProductCategory :one
SELECT id, category, subcategory
FROM product_category
WHERE id = ?
LIMIT 1
`

func (q *Queries) GetProductCategory(ctx context.Context, id int64) (ProductCategory, error) {
	row := q.db.QueryRowContext(ctx, getProductCategory, id)
	var i ProductCategory
	err := row.Scan(&i.ID, &i.Category, &i.Subcategory)
	return i, err
}