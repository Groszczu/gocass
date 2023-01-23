package users

import (
	"github.com/Groszczu/gocass/internal/models"
	"github.com/Groszczu/gocass/internal/repository"
	"github.com/scylladb/gocqlx/v2"
)

type UserRepository = repository.Repository[models.UsersStruct]

func newUserRepository(session *gocqlx.Session) UserRepository {
    return repository.New[models.UsersStruct](session, models.Users)
}
