package repository

import (
	"github.com/vizucode/e-wallet-system/internal/app/dto/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetAllUsers() ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Select("id", "name").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
