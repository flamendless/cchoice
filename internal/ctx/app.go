package ctx

import "cchoice/internal/database"

type App struct {
	DB      database.Service
	Metrics *Metrics
}
