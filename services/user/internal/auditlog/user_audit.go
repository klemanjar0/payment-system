package auditlog

import (
	"context"

	pkgauditlog "github.com/klemanjar0/payment-system/pkg/auditlog"
)

const (
	ActionUserCreated    = "user.created"
	ActionLoginSuccess   = "user.login.success"
	ActionLoginFailure   = "user.login.failure"
	ActionPasswordChange = "user.password.changed"
	ActionTokenRevoked   = "user.token.revoked"
)

const ActorSystem = "system"

type UserAuditLogger struct {
	log pkgauditlog.AuditLogger
}

func New(log pkgauditlog.AuditLogger) *UserAuditLogger {
	return &UserAuditLogger{log: log}
}

func (l *UserAuditLogger) LogUserCreated(ctx context.Context, userID, email string) {
	event := pkgauditlog.NewEvent(ActionUserCreated, ActorSystem, userID, pkgauditlog.StatusSuccess)
	event.Metadata["email"] = email
	l.log.Log(ctx, event)
}

func (l *UserAuditLogger) LogLoginSuccess(ctx context.Context, userID, email string) {
	event := pkgauditlog.NewEvent(ActionLoginSuccess, userID, userID, pkgauditlog.StatusSuccess)
	event.Metadata["email"] = email
	l.log.Log(ctx, event)
}

func (l *UserAuditLogger) LogLoginFailure(ctx context.Context, email, reason string) {
	event := pkgauditlog.NewEvent(ActionLoginFailure, ActorSystem, email, pkgauditlog.StatusFailure)
	event.Error = reason
	event.Metadata["email"] = email
	l.log.Log(ctx, event)
}

func (l *UserAuditLogger) LogPasswordChanged(ctx context.Context, userID string) {
	event := pkgauditlog.NewEvent(ActionPasswordChange, userID, userID, pkgauditlog.StatusSuccess)
	l.log.Log(ctx, event)
}

func (l *UserAuditLogger) LogPasswordChangeFailed(ctx context.Context, userID, reason string) {
	event := pkgauditlog.NewEvent(ActionPasswordChange, userID, userID, pkgauditlog.StatusFailure)
	event.Error = reason
	l.log.Log(ctx, event)
}

func (l *UserAuditLogger) LogTokenRevoked(ctx context.Context, userID, reason string) {
	event := pkgauditlog.NewEvent(ActionTokenRevoked, ActorSystem, userID, pkgauditlog.StatusSuccess)
	event.Metadata["reason"] = reason
	l.log.Log(ctx, event)
}
