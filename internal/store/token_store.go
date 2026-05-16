package store

type TokenStore interface {
	GetUserAccessToken(userID int64) (string, bool)
	SetUserAccessToken(userID int64, accessToken string) error
	DeleteUserAccessToken(userID int64) (bool, error)
	CountUserAccessTokens() int
}
