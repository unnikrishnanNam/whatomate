package handlers

import (
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shridarpatil/whatomate/internal/config"
	"github.com/shridarpatil/whatomate/pkg/whatsapp"
	"github.com/zerodha/fastglue"
	"github.com/zerodha/logf"
	"gorm.io/gorm"
)

// App holds all dependencies for handlers
type App struct {
	Config   *config.Config
	DB       *gorm.DB
	Redis    *redis.Client
	Log      logf.Logger
	WhatsApp *whatsapp.Client
}

// getOrgIDFromContext extracts organization ID from request context (set by auth middleware)
func (a *App) getOrgIDFromContext(r *fastglue.Request) (uuid.UUID, error) {
	orgIDVal := r.RequestCtx.UserValue("organization_id")
	if orgIDVal == nil {
		return uuid.Nil, errors.New("organization_id not found in context")
	}
	// The middleware stores uuid.UUID directly, not as string
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("organization_id is not a valid UUID")
	}
	return orgID, nil
}

// HealthCheck returns server health status
func (a *App) HealthCheck(r *fastglue.Request) error {
	return r.SendEnvelope(map[string]string{
		"status":  "ok",
		"service": "whatomate",
	})
}

// ReadyCheck returns server readiness status
func (a *App) ReadyCheck(r *fastglue.Request) error {
	// Check database connection
	sqlDB, err := a.DB.DB()
	if err != nil {
		return r.SendErrorEnvelope(500, "Database connection error", nil, "")
	}
	if err := sqlDB.Ping(); err != nil {
		return r.SendErrorEnvelope(500, "Database ping failed", nil, "")
	}

	// Check Redis connection
	if err := a.Redis.Ping(r.RequestCtx).Err(); err != nil {
		return r.SendErrorEnvelope(500, "Redis connection error", nil, "")
	}

	return r.SendEnvelope(map[string]string{
		"status": "ready",
	})
}
