
-- +goose Up
-- +goose StatementBegin
CREATE VIRTUAL TABLE tbl_product_categories_fts
USING FTS5(
    category,
    subcategory,
    content='tbl_product_categories',
    content_rowid='id'
);

CREATE TRIGGER tbl_product_categories_after_insert AFTER INSERT ON tbl_product_categories
BEGIN
	INSERT INTO tbl_product_categories_fts(rowid, category, subcategory)
	VALUES (new.id, new.category, new.subcategory);
END;

CREATE TRIGGER tbl_product_categories_after_update AFTER UPDATE ON tbl_product_categories
BEGIN
	UPDATE tbl_product_categories_fts
	SET category = new.category, subcategory = new.subcategory
	WHERE rowid = new.id;
END;

CREATE TRIGGER tbl_product_categories_after_delete AFTER DELETE ON tbl_product_categories
BEGIN
	DELETE FROM tbl_product_categories_fts WHERE rowid = old.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS tbl_product_category_after_delete;
DROP TRIGGER IF EXISTS tbl_product_category_after_update;
DROP TRIGGER IF EXISTS tbl_product_category_after_insert;
DROP TABLE IF EXISTS tbl_product_categories_fts;
