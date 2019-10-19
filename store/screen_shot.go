package store

type ScreenshotMetadata struct {
	ID      string `json:"id" bson:"_id"`
	Url     string `json:"url" bson:"url"`
	Format  string `json:"format"`
	Quality int    `json:"quality"`
	Version int    `json:"version" bson:"version"`
	FileID  string `json:"file_id" bson:"file_id"`
}
