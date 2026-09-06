package auth

type registerResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type loginResponse struct {
	Token string `json:"token"`
}
