package models

type PostRequest struct {
	URL string `json:"url"`
}

type PostResponse struct {
	Result string `json:"result"`
}
