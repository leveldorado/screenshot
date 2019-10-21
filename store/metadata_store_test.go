package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testDatabaseEnvVariable = "SCREENSHOT_TEST_DATABASE"
)

func TestMongodbMetadataRepo_Save(t *testing.T) {
	address := os.Getenv(testDatabaseEnvVariable)
	cl, err := BuildMongoClient(context.Background(), address)
	require.NoError(t, err)
	repo := NewMongodbMetadataRepo(cl, "test", "metadata", "versions")
	require.NoError(t, repo.EnsureIndexes(context.Background()))
	url := uuid.New().String()
	doc := Metadata{ID: uuid.New().String(), CreatedAt: time.Now().UTC().Truncate(time.Millisecond), FileID: uuid.New().String(), Url: url}
	require.NoError(t, repo.Save(context.Background(), &doc))
	require.Equal(t, 1, doc.Version)
	fromDB, err := repo.Get(context.Background(), doc.Url, doc.Version)
	require.NoError(t, err)
	require.Equal(t, doc, fromDB)
	doc2 := Metadata{ID: uuid.New().String(), CreatedAt: time.Now().UTC().Truncate(time.Millisecond), FileID: uuid.New().String(), Url: url}
	require.NoError(t, repo.Save(context.Background(), &doc2))
	require.Equal(t, 2, doc2.Version)

	anotherDoc := Metadata{ID: uuid.New().String(), CreatedAt: time.Now().UTC().Truncate(time.Millisecond), FileID: uuid.New().String(), Url: uuid.New().String()}
	require.NoError(t, repo.Save(context.Background(), &anotherDoc))
	require.Equal(t, 1, doc.Version)

	list, err := repo.GetAllVersions(context.Background(), url)
	require.NoError(t, err)
	require.Contains(t, list, doc)
	require.Contains(t, list, doc2)
	require.NotContains(t, list, anotherDoc)
}
