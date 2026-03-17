-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_staff_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    staff_id INTEGER NOT NULL,
    role TEXT NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')),
    FOREIGN KEY (staff_id) REFERENCES tbl_staffs(id)
);
CREATE INDEX IF NOT EXISTS idx_staff_roles_staff_id ON tbl_staff_roles(staff_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_staff_roles;
-- +goose StatementEnd
