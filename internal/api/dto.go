package api

type shortenRequest struct {
	OriginalUrl string `json:"original_url" binding:"required"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}
