package geocloud

type CreateResponse struct {
	ID string `json:"id"`
}

type StatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
