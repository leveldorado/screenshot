package store

import (
	"context"
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

type MongodbGridFSFileRepo struct {
	bucket *gridfs.Bucket
}

func NewMongodbGridFSFileRepo(ctx context.Context, cl *mongo.Client, databaseName string) (*MongodbGridFSFileRepo, error) {
	bucket, err := gridfs.NewBucket(cl.Database(databaseName))
	if err != nil {
		return nil, fmt.Errorf(`failed to create gridfs bucket: [database: %s, error: %w]`, databaseName, err)
	}
	return &MongodbGridFSFileRepo{bucket: bucket}, nil
}

func (m *MongodbGridFSFileRepo) Save(ctx context.Context, file io.Reader, fileID, filename string) error {
	st, err := m.bucket.OpenUploadStreamWithID(fileID, filename)
	if err != nil {
		return fmt.Errorf(`failed to open upload stream: [file_id: %s, filename: %s, error: %w]"`, fileID, filename, err)
	}
	if _, err = io.Copy(st, file); err != nil {
		return fmt.Errorf(`failed to copy file to gridfs upload stream: [error: %w]`, err)
	}
	st.Close()
	return nil
}

func (m *MongodbGridFSFileRepo) Get(ctx context.Context, fileID string) (io.ReadCloser, error) {
	st, err := m.bucket.OpenDownloadStream(fileID)
	if err != nil {
		return nil, fmt.Errorf(`failed to open download stream: [file_id: %s, error: %w]`, fileID, err)
	}
	return st, nil
}
