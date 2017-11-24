package utils

import (
	"context"
	"errors"

	bolt "github.com/coreos/bbolt"
)

// CheckUserNameExists Checks if the given username exists in database.
// The calling function is responsible to close the DB connection!
func CheckUserNameExists(userName string, db *bolt.DB) bool {
	userNameFound := false
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Users"))
		entry := b.Get([]byte(userName))
		if entry != nil {
			// Whoops! Username is already taken. Set flag accordingly.
			userNameFound = true
			// Stop execution
			return nil
		}
		return nil
	})
	return userNameFound
}

// https://medium.com/@matryer/context-keys-in-go-5312346a868d
type contextKey string

func (c contextKey) String() string {
	return "gamecharacters context key " + string(c)
}

var (
	contextKeyAuthenticated = contextKey("authenticated")
	contextKeyAuthUserName  = contextKey("username")
)

// AuthData holds auth data from context
type AuthData struct {
	UserName      string
	Authenticated bool
}

// PutContextAuthData puts auth data into a context
func PutContextAuthData(ctx context.Context, authenticated bool, userName string) context.Context {
	contextWithValues := context.WithValue(ctx, contextKeyAuthUserName, userName)
	contextWithValues = context.WithValue(contextWithValues, contextKeyAuthenticated, authenticated)
	return contextWithValues
}

// GetContextAuthData gets AuthData from a context object
func GetContextAuthData(ctx context.Context) (AuthData, error) {
	authenticated, ok := ctx.Value(contextKeyAuthenticated).(bool)
	if !ok {
		return AuthData{}, errors.New("Error getting AuthData")
	}
	userName, ok := ctx.Value(contextKeyAuthUserName).(string)
	if !ok {
		return AuthData{}, errors.New("Error getting AuthData")
	}
	return AuthData{Authenticated: authenticated, UserName: userName}, nil
}
