package users

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	errors2 "github.com/cloudradar-monitoring/rport/server/api/errors"
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
	Delete(usernameToDelete string) error
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

func (as *APIService) Change(usr *User, userKey string) error {
	if usr.Password != "" {
		passHash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		usr.Password = strings.Replace(string(passHash), htpasswdBcryptAltPrefix, htpasswdBcryptPrefix, 1)
	}

	err := as.validate(usr, userKey)
	if err != nil {
		return err
	}

	if as.ProviderType == ProviderFromFile {
		return as.changeUserInFile(usr, userKey)
	}

	if as.ProviderType == ProviderFromDB {
		return as.changeUserInDB(usr, userKey)
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
		if (dataToChange.Username == "" || dataToChange.Username == userKeyToFind) && dataToChange.Password == "" && len(dataToChange.Groups) == 0 {
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

func (as *APIService) addUserToDB(dataToChange *User) error {
	if dataToChange.Username != "" {
		existingUser, err := as.DB.GetByUsername(dataToChange.Username)
		if err != nil {
			return err
		}
		if existingUser != nil {
			return errors2.APIError{
				Message: "Another user with this username already exists",
				Code:    http.StatusBadRequest,
			}
		}
	}

	err := as.DB.Add(dataToChange)
	if err != nil {
		return err
	}

	return nil
}

func (as *APIService) updateUserInDB(dataToChange *User, userKeyToFind string) error {
	existingUser, err := as.DB.GetByUsername(userKeyToFind)
	if err != nil {
		return err
	}

	if existingUser == nil || existingUser.Username == "" {
		return errors2.APIError{
			Message: fmt.Sprintf("cannot find user by username '%s'", userKeyToFind),
			Code:    http.StatusNotFound,
		}
	}

	if dataToChange.Username != "" && dataToChange.Username != userKeyToFind {
		existingUser, err := as.DB.GetByUsername(dataToChange.Username)
		if err != nil {
			return err
		}
		if existingUser != nil {
			return errors2.APIError{
				Message: "Another user with this username already exists",
				Code:    http.StatusBadRequest,
			}
		}
	}

	err = as.DB.Update(dataToChange, userKeyToFind)
	if err != nil {
		return err
	}
	return nil
}

func (as *APIService) changeUserInDB(dataToChange *User, userKeyToFind string) error {
	if userKeyToFind == "" {
		return as.addUserToDB(dataToChange)
	}
	return as.updateUserInDB(dataToChange, userKeyToFind)
}

func (as *APIService) addUserToFile(dataToChange *User) error {
	users, err := as.FileProvider.ReadUsersFromFile()
	if err != nil {
		return err
	}

	for i := range users {
		if users[i].Username == dataToChange.Username {
			return errors2.APIError{
				Message: "Another user with this username already exists",
				Code:    http.StatusBadRequest,
			}
		}
	}

	users = append(users, dataToChange)
	err = as.FileProvider.SaveUsersToFile(users)
	if err != nil {
		return err
	}
	return nil
}

func (as *APIService) updateUserInFile(dataToChange *User, userKeyToFind string) error {
	users, err := as.FileProvider.ReadUsersFromFile()
	if err != nil {
		return err
	}

	userFound := false
	for i := range users {
		if users[i].Username == userKeyToFind {
			userFound = true
		}
		if dataToChange.Username != "" && users[i].Username == dataToChange.Username && dataToChange.Username != userKeyToFind {
			return errors2.APIError{
				Message: "Another user with this username already exists",
				Code:    http.StatusBadRequest,
			}
		}
	}

	if !userFound {
		return errors2.APIError{
			Message: fmt.Sprintf("cannot find user by username '%s'", userKeyToFind),
			Code:    http.StatusNotFound,
		}
	}

	for i := range users {
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

func (as *APIService) changeUserInFile(dataToChange *User, userKeyToFind string) error {
	if userKeyToFind != "" {
		return as.updateUserInFile(dataToChange, userKeyToFind)
	}

	return as.addUserToFile(dataToChange)
}

func (as *APIService) Delete(usernameToDelete string) error {
	if as.ProviderType == ProviderFromFile {
		return as.deleteUserFromFile(usernameToDelete)
	}

	if as.ProviderType == ProviderFromDB {
		return as.deleteUserFromDB(usernameToDelete)
	}

	return fmt.Errorf("unknown user data provider type: %d", as.ProviderType)
}

func (as *APIService) deleteUserFromDB(usernameToDelete string) error {
	user, err := as.DB.GetByUsername(usernameToDelete)
	if err != nil {
		return err
	}

	if user == nil || user.Username == "" {
		return errors2.APIError{
			Message: fmt.Sprintf("cannot find user by username '%s'", usernameToDelete),
			Code:    http.StatusNotFound,
		}
	}

	return as.DB.Delete(usernameToDelete)
}

func (as *APIService) deleteUserFromFile(usernameToDelete string) error {
	usersFromFile, err := as.FileProvider.ReadUsersFromFile()
	if err != nil {
		return err
	}
	foundIndex := -1
	for i := range usersFromFile {
		if usersFromFile[i].Username == usernameToDelete {
			foundIndex = i
			break
		}
	}

	if foundIndex < 0 {
		return errors2.APIError{
			Message: fmt.Sprintf("cannot find user by username '%s'", usernameToDelete),
			Code:    http.StatusNotFound,
		}
	}

	usersToWriteToFile := append(usersFromFile[:foundIndex], usersFromFile[foundIndex+1:]...)
	err = as.FileProvider.SaveUsersToFile(usersToWriteToFile)
	if err != nil {
		return err
	}
	return nil
}
