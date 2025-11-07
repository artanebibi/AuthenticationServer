package service

import (
	"AuthServer/internal/domain/models"
	"AuthServer/internal/repository"
)

type IUserService interface {
	//FindAll() []models.User
	FindById(Id string) (*models.User, error)
	FindByEmailOrUsername(identifier string) (*models.User, error)
	Save(models.User) models.User
	Update(models.User) error
	Delete(id string) error
}

type UserService struct {
	userRepository repository.IUserRepository
}

func NewUserService(repo repository.IUserRepository) *UserService {
	return &UserService{
		userRepository: repo,
	}
}

//func (u *UserService) FindAll() []models.User {
//	return u.userRepository.FindAll()
//}

func (u *UserService) FindById(Id string) (*models.User, error) {
	user, err := u.userRepository.FindById(Id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) FindByEmailOrUsername(identifier string) (*models.User, error) {
	return u.userRepository.FindByEmailOrUsername(identifier)
}

func (u *UserService) Save(user models.User) models.User {
	u.userRepository.Save(user)
	return user
}

func (u *UserService) Update(user models.User) error {
	return u.userRepository.Update(user)
}

func (u *UserService) Delete(id string) error {
	return u.userRepository.Delete(id)
}
