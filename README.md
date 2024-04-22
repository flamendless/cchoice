# Setup

- `sqlc generate`
- `sql-migrate up`

# Usage

- Sample `go run ./main.go parse_xlsx -p "xlsx/sample.xlsx" -s "Sheet" -t "sample" --use_db --db_path "file:./test.db"`
- Sample strict `go run ./main.go parse_xlsx -p "xlsx/sample.xlsx" -s "Sheet" -t "sample" -x 1 --use_db --db_path "file:./test.db"`
