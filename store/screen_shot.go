package store

type ScreenshotMetadata struct {
	ID      string `json:"id" bson:"_id"`
	Url     string `json:"url" bson:"url"`
	Version int    `json:"version" bson:"version"`
}
