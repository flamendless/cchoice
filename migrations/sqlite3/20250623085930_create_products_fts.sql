
-- +goose Up
CREATE VIRTUAL TABLE tbl_products_fts
USING fts5(
    serial,
    name,
    content='tbl_products',
    content_rowid='id'
);

-- -- +goose StatementBegin
CREATE TRIGGER tbl_products_after_insert AFTER INSERT ON tbl_products
BEGIN
	INSERT INTO tbl_products_fts(rowid, serial, name)
	VALUES (new.id, new.serial, new.name);
END;

CREATE TRIGGER tbl_products_after_update AFTER UPDATE ON tbl_products
BEGIN
	UPDATE tbl_products_fts
	SET serial = new.serial, name = new.name
	WHERE rowid = new.id;
END;

CREATE TRIGGER tbl_products_after_delete AFTER DELETE ON tbl_products
BEGIN
	DELETE FROM tbl_products_fts WHERE rowid = old.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER tbl_products_after_delete;
DROP TRIGGER tbl_products_after_update;
DROP TRIGGER tbl_products_after_insert;
DROP TABLE tbl_products_fts;
