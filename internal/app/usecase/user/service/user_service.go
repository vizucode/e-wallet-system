package service

import (
	"github.com/vizucode/e-wallet-system/internal/app/dto/domains"
	"github.com/vizucode/e-wallet-system/internal/app/usecase/user/repository"
)

type UserService interface {
	GetAllUsers() ([]domains.GetUsersResponse, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetAllUsers() ([]domains.GetUsersResponse, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	var response []domains.GetUsersResponse
	for _, user := range users {
		response = append(response, domains.GetUsersResponse{
			UserID: user.ID,
			Name:   user.Name,
		})
	}

	return response, nil
}
