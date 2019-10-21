package store

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

func TestMongodbGridFSFileRepo_Save(t *testing.T) {
	address := os.Getenv(testDatabaseEnvVariable)
	cl, err := BuildMongoClient(context.Background(), address)
	require.NoError(t, err)
	repo, err := NewMongodbGridFSFileRepo(context.Background(), cl, "test")
	require.NoError(t, err)
	data := uuid.New().String()
	file := strings.NewReader(data)
	fileID := uuid.New().String()
	require.NoError(t, repo.Save(context.Background(), file, fileID, fileID))
	fileFromDb, err := repo.Get(context.Background(), fileID)
	require.NoError(t, err)
	defer fileFromDb.Close()
	dataFromDB, err := ioutil.ReadAll(fileFromDb)
	require.Equal(t, data, string(dataFromDB))
}
