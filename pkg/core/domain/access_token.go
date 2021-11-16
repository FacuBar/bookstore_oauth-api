package domain

type AccessToken struct {
	AccessToken string `json:"access_token"`
	Expires     int64  `json:"expires"`
	UserId      int64  `json:"user_id"`
}
