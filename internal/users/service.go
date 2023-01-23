package users

import (
	"github.com/Groszczu/gocass/internal/models"
	"github.com/scylladb/gocqlx/v2"
)

type Service struct {
	userRepo UserRepository
}

func NewService(session *gocqlx.Session) Service {
	return Service{
		newUserRepository(session),
	}
}

func (s Service) RegisterUser(user *models.UsersStruct) error {
	if user.Id != models.EmptyUUID() {
		return s.userRepo.Insert(user)
	}
	user.Id = models.RandomUUID()
	return s.userRepo.Insert(user)
}

func (s Service) GetUser(user *models.UsersStruct) error {
	return s.userRepo.GetOne(user)
}

func (s Service) FindUsers(user *models.UsersStruct) (*[]models.UsersStruct, error) {
	return s.userRepo.GetAll(user)
}
