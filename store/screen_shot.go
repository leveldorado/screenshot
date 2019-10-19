package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type ScreenshotMetadata struct {
	ID        string    `json:"id" bson:"_id"`
	Url       string    `json:"url" bson:"url"`
	Format    string    `json:"format"`
	Quality   int       `json:"quality"`
	Version   int       `json:"version" bson:"version"`
	FileID    string    `json:"file_id" bson:"file_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type MongodbScreenshotMetadataRepo struct {
	db                       *mongo.Database
	metadataCollection       string
	versionCounterCollection string
}

func NewMongodbScreenshotMetadataRepo(cl *mongo.Client, database, metadataCollection, versionCounterCollection string) *MongodbScreenshotMetadataRepo {
	return &MongodbScreenshotMetadataRepo{db: cl.Database(database), metadataCollection: metadataCollection, versionCounterCollection: versionCounterCollection}
}

func (m *MongodbScreenshotMetadataRepo) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{{
		Keys: bson.D{{
			Name:  "url",
			Value: 1,
		}},
	}}
	if _, err := m.db.Collection(m.metadataCollection).Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf(`failed to create metadata indexes: [indexes: %+v, error: %w]`, indexes, err)
	}
	return nil
}

func (m *MongodbScreenshotMetadataRepo) Save(ctx context.Context, doc *ScreenshotMetadata) error {
	versionDoc := struct {
		ID      string `bson:"_id"`
		Version int    `bson:"version"`
	}{}
	if err := m.db.Collection(m.versionCounterCollection).
		FindOneAndUpdate(ctx, bson.M{"_id": doc.Url}, bson.M{"$inc": bson.M{"version": 1}}).
		Decode(versionDoc); err != nil {
		return fmt.Errorf(`failed to generete new version: [id: %s, error: %w]`, doc.Url, err)
	}
	doc.Version = versionDoc.Version
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}
	_, err := m.db.Collection(m.metadataCollection).InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf(`failed to insert doc to metadata collection: [doc: %+v, error: %w]`, doc, err)
	}
	return nil
}
