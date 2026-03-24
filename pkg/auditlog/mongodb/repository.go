package mongodb

import (
	"context"
	"fmt"

	"github.com/klemanjar0/payment-system/pkg/auditlog"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const defaultCollection = "audit_logs"

// Repository persists audit events to a MongoDB collection.
type Repository struct {
	collection *mongo.Collection
}

// New creates a Repository targeting the given database and collection.
// If collectionName is empty, "audit_logs" is used.
func New(client *mongo.Client, database, collectionName string) *Repository {
	if collectionName == "" {
		collectionName = defaultCollection
	}
	return &Repository{
		collection: client.Database(database).Collection(collectionName),
	}
}

// Save inserts an audit event document into MongoDB.
func (r *Repository) Save(ctx context.Context, event auditlog.Event) error {
	_, err := r.collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("auditlog/mongodb: insert failed: %w", err)
	}
	return nil
}

// EnsureIndexes creates indexes on the audit_logs collection for
// efficient querying by service/action, target user, and time range.
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: map[string]int{"service": 1, "action": 1, "timestamp": -1},
		},
		{
			Keys: map[string]int{"target_id": 1, "timestamp": -1},
		},
		{
			Keys: map[string]int{"actor_id": 1, "timestamp": -1},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes, options.CreateIndexes())
	if err != nil {
		return fmt.Errorf("auditlog/mongodb: failed to create indexes: %w", err)
	}
	return nil
}

// compile-time check that Repository implements auditlog.Repository
var _ auditlog.Repository = (*Repository)(nil)
