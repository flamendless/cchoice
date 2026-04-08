-- +goose Up
-- Create holidays table
CREATE TABLE IF NOT EXISTS tbl_holidays (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_holidays_date ON tbl_holidays(date);

-- Add holiday columns to staff attendances
ALTER TABLE tbl_staff_attendances ADD COLUMN is_holiday INTEGER DEFAULT 0;
ALTER TABLE tbl_staff_attendances ADD COLUMN holiday_type TEXT;
ALTER TABLE tbl_staff_attendances ADD COLUMN holiday_name TEXT;

-- +goose Down
ALTER TABLE tbl_staff_attendances DROP COLUMN is_holiday;
ALTER TABLE tbl_staff_attendances DROP COLUMN holiday_type;
ALTER TABLE tbl_staff_attendances DROP COLUMN holiday_name;
DROP INDEX idx_holidays_date;
DROP TABLE tbl_holidays;
