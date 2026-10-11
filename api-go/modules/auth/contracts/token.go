package contracts

type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
	ExpiresAt   string `json:"expiresAt"`
}
