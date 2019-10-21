package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

func BuildMongoClient(ctx context.Context, addr string) (*mongo.Client, error) {
	opt := options.Client().ApplyURI(addr)
	if err := opt.Validate(); err != nil {
		return nil, fmt.Errorf(`invalid url: [url: %s, error: %w]`, addr, err)
	}
	cl, err := mongo.NewClient(opt)
	if err != nil {
		return nil, fmt.Errorf(`failed to create client: [opt: %+v, error: %w]`, opt, err)
	}
	if err := cl.Connect(ctx); err != nil {
		return nil, fmt.Errorf(`failed to connect mongodb: [error: %w]`, err)
	}
	return cl, nil
}

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
		Keys: bson.M{"url": 1},
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
	f := bson.M{"_id": doc.Url}
	u := bson.M{"$inc": bson.M{"version": 1}}
	t := true
	after := options.After
	opt := &options.FindOneAndUpdateOptions{Upsert: &t, ReturnDocument: &after}
	if err := m.db.Collection(m.versionCounterCollection).FindOneAndUpdate(ctx, f, u, opt).Decode(&versionDoc); err != nil {
		return fmt.Errorf(`failed to generete new version: [filter: %v, update: %v, error: %w]`, f, u, err)
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
