package storage

import (
	"encoding/json"
	"github.com/jlisicki/middlewarebuilder"
	"io"
)

type (
	UserID string
	// User entity
	User struct {
		ID   UserID
		Name string
	}

	// UserRepository is an alias of a long generic syntax of user repository type
	UserRepository = Repository[User, UserID]

	userIDSerializer struct{}
	userSerializer   struct{}
)

func (u User) Identifier() UserID {
	return u.ID
}

func (u userIDSerializer) Serialize(t UserID) ([]byte, error) {
	return []byte(t), nil
}

func (u userIDSerializer) UnSerialize(bytes []byte) (UserID, error) {
	return UserID(bytes), nil
}

func (u userSerializer) Serialize(t User) ([]byte, error) {
	return json.Marshal(t)
}

func (u userSerializer) UnSerialize(bytes []byte) (User, error) {
	var user User
	err := json.Unmarshal(bytes, &user)
	return user, err
}

func NewUserRepository(debugWriter io.Writer) (UserRepository, error) {
	builder := middlewarebuilder.NewBuilder[UserRepository]()
	return builder.
		Add(middlewarebuilder.FactoryFunc[UserRepository](func(next UserRepository) (UserRepository, error) {
			return Telemetry[User, UserID]{Next: next}, nil
		})).
		Add(middlewarebuilder.FactoryFunc[UserRepository](func(next UserRepository) (UserRepository, error) {
			return Debug[User, UserID]{Next: next, Output: debugWriter, Label: "CacheCall"}, nil
		})).
		Add(middlewarebuilder.FactoryFunc[UserRepository](func(next UserRepository) (UserRepository, error) {
			return &Cache[User, UserID]{Next: next, cached: make(map[UserID]User)}, nil
		})).
		Add(middlewarebuilder.FactoryFunc[UserRepository](func(next UserRepository) (UserRepository, error) {
			return Debug[User, UserID]{Next: next, Output: debugWriter, Label: "StorageCall"}, nil
		})).
		WithHandler(NewInMemoryRepository[User, UserID](userIDSerializer{}, userSerializer{})).Build()
}
