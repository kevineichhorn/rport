package users

import (
	"context"
	"fmt"
	errors2 "github.com/cloudradar-monitoring/rport/server/api/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

type ProviderType int

const (
	ProviderFromFile ProviderType = iota + 1
	ProviderFromStaticPassword
	ProviderFromDB
)

type DatabaseProvider interface {
	GetAll() ([]*User, error)
	GetByUsername(username string) (*User, error)
	Add(usr *User) error
	Update(usr *User, usernameToUpdate string) error
}

type FileProvider interface {
	ReadUsersFromFile() ([]*User, error)
	SaveUsersToFile(users []*User) error
}

type APIService struct {
	ProviderType ProviderType
	FileProvider FileProvider
	DB           DatabaseProvider
}

func (as *APIService) GetAll() ([]*User, error) {
	if as.ProviderType == ProviderFromFile {
		authUsers, err := as.FileProvider.ReadUsersFromFile()
		if err != nil {
			return nil, err
		}
		return authUsers, nil
	}

	if as.ProviderType == ProviderFromDB {
		usrs, err := as.DB.GetAll()
		if err != nil {
			return nil, err
		}
		return usrs, nil
	}

	return nil, fmt.Errorf("unknown user data provider type: %d", as.ProviderType)
}

func (as *APIService) Change(ctx context.Context, usr *User, userKey string) error {
	if usr.Password != "" {
		passHash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		usr.Password = string(passHash)
	}

	err := as.validate(usr, userKey)
	if err != nil {
		return err
	}

	if as.ProviderType == ProviderFromFile {
		return as.addUserToFile(usr, userKey)
	}

	if as.ProviderType == ProviderFromDB {
		return as.addUserToDB(usr, userKey)
	}

	return fmt.Errorf("unknown user data provider type: %d", as.ProviderType)
}

func (as *APIService) validate(dataToChange *User, userKeyToFind string) error {
	errs := []string{}

	if userKeyToFind == "" {
		if dataToChange.Username == "" {
			errs = append(errs, "username is required")
		}
		if dataToChange.Password == "" {
			errs = append(errs, "password is required")
		}
	} else {
		if dataToChange.Username == "" && dataToChange.Password == "" && len(dataToChange.Groups) == 0 {
			errs = append(errs, "nothing to change")
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors2.APIError{
		Message: strings.Join(errs, ", "),
		Code:    http.StatusBadRequest,
	}
}

func (as *APIService) addUserToDB(dataToChange *User, userKeyToFind string) error {
	if dataToChange.Username != "" {
		existingUser, err := as.DB.GetByUsername(dataToChange.Username)
		if err != nil {
			return err
		}
		if existingUser != nil && (userKeyToFind == "" || existingUser.Username != userKeyToFind) {
			return errors2.APIError{
				Message: "Another user with this username already exists",
				Code:    http.StatusBadRequest,
			}
		}
	}

	if userKeyToFind != "" {
		err := as.DB.Update(dataToChange, userKeyToFind)
		if err != nil {
			return err
		}
	}

	err := as.DB.Add(dataToChange)
	if err != nil {
		return err
	}

	return nil
}

func (as *APIService) addUserToFile(dataToChange *User, userKeyToFind string) error {
	users, err := as.FileProvider.ReadUsersFromFile()
	if err != nil {
		return err
	}

	for i := range users {
		if users[i].Username == dataToChange.Username &&
			(dataToChange.Username != userKeyToFind || userKeyToFind == "") {
			return errors2.APIError{
				Message: "Another user with this username already exists",
				Code:    http.StatusBadRequest,
			}
		}
	}

	if userKeyToFind == "" {
		users = append(users, dataToChange)
		err := as.FileProvider.SaveUsersToFile(users)
		if err != nil {
			return err
		}
	}

	for i := range users {
		if users[i].Username != userKeyToFind {
			continue
		}
		if dataToChange.Password != "" {
			users[i].Password = dataToChange.Password
		}
		if dataToChange.Groups != nil && len(dataToChange.Groups) > 0 {
			users[i].Groups = dataToChange.Groups
		}
		if dataToChange.Username != "" {
			users[i].Username = dataToChange.Username
		}
	}

	err = as.FileProvider.SaveUsersToFile(users)
	if err != nil {
		return err
	}

	return nil
}
