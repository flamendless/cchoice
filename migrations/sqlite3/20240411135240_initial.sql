-- +goose NO TRANSACTION

-- +goose Up
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA mmap_size = 134217728;
PRAGMA journal_size_limit = 27103364;
PRAGMA cache_size = 2000;

-- +goose Down
