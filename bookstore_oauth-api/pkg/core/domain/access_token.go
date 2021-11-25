package domain

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`

	Expires int64 `json:"expires"`

	UserId   int64  `json:"user_id"`
	UserRole string `json:"user_role"`
}
