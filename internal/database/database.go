package database

import (
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type DBMode string

const (
	DB_MODE_RO DBMode = "ro"
	DB_MODE_RW DBMode = "rw"
)

type Service interface {
	Health() map[string]string
	Close() error
	GetQueries() *queries.Queries
}

type service struct {
	db      *sql.DB
	queries *queries.Queries
}

func (s *service) GetQueries() *queries.Queries {
	return s.queries
}

var (
	dburl        = os.Getenv("DB_URL")
	dbInstanceRO *service
	dbInstanceRW *service
)

func New(mode DBMode) Service {
	if dburl == "" {
		panic(fmt.Errorf("%w. DB_URL", errs.ERR_ENV_VAR_REQUIRED))
	}

	switch mode {
	case DB_MODE_RO:
		if dbInstanceRO != nil {
			return dbInstanceRO
		}
	case DB_MODE_RW:
		if dbInstanceRW != nil {
			return dbInstanceRW
		}
	default:
		panic("db mode enum not handled")
	}

	logs.Log().Info("Initializing DB...", zap.String("mode", string(mode)))
	dataSourceName := dburl + "?_journal_mode=wal" + "&mode=" + string(mode)
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		logs.Log().Fatal("db open", zap.Error(err))
	}

	switch mode {
	case DB_MODE_RO:
		dbInstanceRO = &service{
			db:      db,
			queries: queries.New(db),
		}
		return dbInstanceRO
	case DB_MODE_RW:
		dbInstanceRW = &service{
			db:      db,
			queries: queries.New(db),
		}
		return dbInstanceRW
	default:
		panic("db mode enum not handled")
	}
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		logs.Log().Fatal("db down", zap.Error(err))
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *service) Close() error {
	logs.Log().Info("Disconnected from database", zap.String("db url", dburl))
	return s.db.Close()
}

var _ Service = (*service)(nil)
