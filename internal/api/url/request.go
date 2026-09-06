package url

type shortenRequest struct {
	OriginalUrl string `json:"original_url" binding:"required"`
}
