package auditlog

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/klemanjar0/payment-system/pkg/logger"
)

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
)

type Event struct {
	ID        string         `json:"id"`
	Service   string         `json:"service"`
	Action    string         `json:"action"`
	ActorID   string         `json:"actor_id"`
	TargetID  string         `json:"target_id"`
	Status    string         `json:"status"`
	Metadata  map[string]any `json:"metadata"`
	Error     string         `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type Repository interface {
	Save(ctx context.Context, event Event) error
}

type AuditLogger interface {
	Log(ctx context.Context, event Event)
}

type Logger struct {
	repo    Repository
	service string
}

func New(repo Repository, service string) *Logger {
	return &Logger{repo: repo, service: service}
}

func (l *Logger) Log(ctx context.Context, event Event) {
	event.ID = uuid.New().String()
	event.Service = l.service
	event.Timestamp = time.Now().UTC()

	l.zapLog(event)

	go func() {
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := l.repo.Save(saveCtx, event); err != nil {
			logger.Error("[AUDIT] failed to persist audit event",
				"action", event.Action,
				"actor", event.ActorID,
				"target", event.TargetID,
				"err", err,
			)
		}
	}()
}

func (l *Logger) zapLog(e Event) {
	keysAndValues := []any{
		"service", e.Service,
		"action", e.Action,
		"actor", e.ActorID,
		"target", e.TargetID,
		"status", e.Status,
	}

	if e.Error != "" {
		keysAndValues = append(keysAndValues, "error", e.Error)
	}

	if e.Status == StatusFailure {
		logger.Warn("[AUDIT]", keysAndValues...)
	} else {
		logger.Info("[AUDIT]", keysAndValues...)
	}
}

func NewEvent(action, actorID, targetID, status string) Event {
	return Event{
		Action:   action,
		ActorID:  actorID,
		TargetID: targetID,
		Status:   status,
		Metadata: make(map[string]any),
	}
}
