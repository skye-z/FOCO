package audit

import (
	"context"
	"encoding/json"
	"time"

	"xorm.io/xorm"
)

type AuditLog struct {
	Id           string    `xorm:"'id' pk uuid"`
	ActorUserId  *string   `xorm:"'actor_user_id' uuid"`
	Action       string    `xorm:"'action' notnull"`
	ResourceType string    `xorm:"'resource_type' notnull"`
	ResourceId   string    `xorm:"'resource_id' notnull"`
	Payload      string    `xorm:"'payload' notnull default '{}'"`
	CreatedAt    time.Time `xorm:"'created_at' notnull default now()"`
}

func (AuditLog) TableName() string { return "audit_logs" }

type XormWriter struct {
	engine *xorm.Engine
}

func NewXormWriter(engine *xorm.Engine) *XormWriter {
	return &XormWriter{engine: engine}
}

func (w *XormWriter) Write(ctx context.Context, actor, action, resourceType, resourceID string, payload map[string]any) error {
	payloadJSON, _ := json.Marshal(payload)
	log := &AuditLog{
		ActorUserId:  nilIfEmpty(actor),
		Action:       action,
		ResourceType: resourceType,
		ResourceId:   resourceID,
		Payload:      string(payloadJSON),
		CreatedAt:    time.Now(),
	}
	_, err := w.engine.Context(ctx).Insert(log)
	return err
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
