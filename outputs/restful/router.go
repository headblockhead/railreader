package restful

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(log *slog.Logger, dbpool *pgxpool.Pool) *gin.Engine {
	r := gin.Default()
	// Define your routes here
	return r
}
