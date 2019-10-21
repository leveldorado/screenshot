package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type ErrNotFound struct{}

func (ErrNotFound) Error() string { return "Not found" }

type Metadata struct {
	ID        string    `json:"id" bson:"_id"`
	Url       string    `json:"url" bson:"url"`
	Format    string    `json:"format"`
	Quality   int       `json:"quality"`
	Version   int       `json:"version" bson:"version"`
	FileID    string    `json:"file_id" bson:"file_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

func (m Metadata) GetContentType() string {
	return fmt.Sprintf(`image/%s`, m.Format)
}

type MetadataByVersionDesc []Metadata

func (list MetadataByVersionDesc) Len() int      { return len(list) }
func (list MetadataByVersionDesc) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
func (list MetadataByVersionDesc) Less(i, j int) bool {
	return list[i].Version > list[j].Version
}

type MongodbMetadataRepo struct {
	db                       *mongo.Database
	metadataCollection       string
	versionCounterCollection string
}

func NewMongodbMetadataRepo(cl *mongo.Client, database, metadataCollection, versionCounterCollection string) *MongodbMetadataRepo {
	return &MongodbMetadataRepo{db: cl.Database(database), metadataCollection: metadataCollection, versionCounterCollection: versionCounterCollection}
}

func (m *MongodbMetadataRepo) EnsureIndexes(ctx context.Context) error {
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

func (m *MongodbMetadataRepo) Save(ctx context.Context, doc *Metadata) error {
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

func (m *MongodbMetadataRepo) Get(ctx context.Context, url string, version int) (Metadata, error) {
	var doc Metadata
	q := bson.M{"url": url, "version": version}
	err := m.db.Collection(m.metadataCollection).FindOne(ctx, q).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Metadata{}, ErrNotFound{}
	}
	if err != nil {
		return Metadata{}, fmt.Errorf(`failed to find document: [q: %+v, collection_name: %s, error: %w]`, q, m.metadataCollection, err)
	}
	return doc, nil
}

func (m *MongodbMetadataRepo) GetAllVersions(ctx context.Context, url string) ([]Metadata, error) {
	var list []Metadata
	q := bson.M{"url": url}
	res, err := m.db.Collection(m.metadataCollection).Find(ctx, q)
	if err != nil {
		return nil, fmt.Errorf(`failed to find documents: [q: %v, collection_name: %s, error: %w]`, q, m.metadataCollection, err)
	}
	if err = res.All(ctx, &list); err != nil {
		return nil, fmt.Errorf(`failed to decode result: [error: %w]`, err)
	}
	return list, nil
}
