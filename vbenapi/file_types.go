package vbenapi

type uploadedFileResponse struct {
	URL     string `json:"url"`
	Storage string `json:"storage"`
	Name    string `json:"name"`
}
