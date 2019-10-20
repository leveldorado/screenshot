package capture

import "github.com/leveldorado/screenshot/store"

type ShotResponse struct {
	Success  bool           `json:"success"`
	Metadata store.Metadata `json:"metadata"`
	Error    string         `json:"error"`
}

type ShotRequest struct {
	URL string `json:"url"`
}

const ShotRequestTopic = "shot_request"
