package repository

import "context"


type UserRepository interface {
	UpdateAvatar(ctx context.Context, userID string, newFileName string) (oldFileName string, err error)
}