package db

import (
	"fmt"
)

type UserRepository interface {
	Create() error
}


type UserRepositoryImpl struct {
	// db *sql.DB
}

func NewUserRespository() UserRepository {
	return &UserRepositoryImpl{
		// db: db
	}
}

func (u *UserRepositoryImpl) Create() error {
	fmt.Println("Creating User in UserRepository")
	return nil
}