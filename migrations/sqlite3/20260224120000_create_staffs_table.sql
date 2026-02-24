-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_staffs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    middle_name TEXT,
    last_name TEXT NOT NULL,
    birthdate TEXT NOT NULL,
    sex TEXT NOT NULL,
    date_hired TEXT NOT NULL,
    time_in_schedule TEXT,
    time_out_schedule TEXT,
    position TEXT NOT NULL,
    user_type TEXT NOT NULL CHECK(user_type IN ('SUPERUSER', 'STAFF')),
    email TEXT NOT NULL UNIQUE,
    mobile_no TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT ('1970-01-01 00:00:00+00:00')
);

CREATE TABLE IF NOT EXISTS tbl_staff_attendances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    staff_id INTEGER NOT NULL,
    for_date TEXT NOT NULL,
    time_in TEXT,
    time_out TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (staff_id) REFERENCES tbl_staffs(id),
    UNIQUE(staff_id, for_date)
);

-- +goose Down
DROP TABLE IF EXISTS tbl_staff_attendances;
DROP TABLE IF EXISTS tbl_staffs;
