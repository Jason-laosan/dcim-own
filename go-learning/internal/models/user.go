package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string    `json:"username" gorm:"not null;uniqueIndex"`
	Email     string    `json:"email" gorm:"not null;uniqueIndex"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(username, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	err = s.db.Create(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByEmail(email string) (*User, error) {
	var user User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByUsername(username string) (*User, error) {
	var user User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByID(id int) (*User, error) {
	var user User
	err := s.db.First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) ValidatePassword(user *User, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return errors.New("invalid password")
	}
	return nil
}
