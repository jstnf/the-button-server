package data

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type UserStorage interface {
	GetUserById(id string) (*User, error)
}

type LocalUserStorage struct {
	users map[string]*User
}

func NewLocalUserStorage() *LocalUserStorage {
	return &LocalUserStorage{
		users: make(map[string]*User),
	}
}

type Users struct {
	Users []User `json:"users"`
}

func (s *LocalUserStorage) Init() error {
	jsonFile, err := os.Open("users.json")
	if err != nil {
		return err
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			panic(err)
		}
	}(jsonFile)

	byteValue, _ := io.ReadAll(jsonFile)
	var users Users
	if err := json.Unmarshal(byteValue, &users); err != nil {
		return err
	}
	for _, user := range users.Users {
		s.users[user.UserId] = &user
	}
	return nil
}

func (s *LocalUserStorage) GetUserById(id string) (*User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}
