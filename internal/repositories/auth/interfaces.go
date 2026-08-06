package auth

import "github.com/projsonal/gowms/internal/model"

type Repository interface {
	SaveRefreshToken(t *model.RefreshToken) error
	FindActiveRefreshToken(userID uint, tokenHash string) (*model.RefreshToken, error)
	RevokeAllUserTokens(userID uint) error

	ListActiveSessions(userID uint) ([]model.RefreshToken, error)
	RevokeSession(userID, sessionID uint) error
}
