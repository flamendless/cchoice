-- +goose Up
-- +goose StatementBegin
UPDATE tbl_products
SET name = TRIM(
    REPLACE(
        REPLACE(name, CHAR(13), ''),  -- remove \r (Windows)
        CHAR(10), ' '                 -- replace \n with space
    )
)
WHERE name LIKE '%' || CHAR(10) || '%'
   OR name LIKE '%' || CHAR(13) || '%';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
