package users

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type DBProviderMock struct {
	UsersToGive   []*User
	UsersToAdd    []*User
	UsersToUpdate []*User
	ErrorToGive   error
}

func (dpm *DBProviderMock) GetAll() ([]*User, error) {
	return dpm.UsersToGive, dpm.ErrorToGive
}

func (dpm *DBProviderMock) GetByUsername(username string) (*User, error) {
	var usr *User
	for i := range dpm.UsersToGive {
		if dpm.UsersToGive[i].Username == username {
			usr = dpm.UsersToGive[i]
		}
	}

	return usr, dpm.ErrorToGive
}

func (dpm *DBProviderMock) Add(usr *User) error {
	if dpm.UsersToAdd == nil {
		dpm.UsersToAdd = []*User{}
	}

	dpm.UsersToAdd = append(dpm.UsersToAdd, usr)

	return dpm.ErrorToGive
}

func (dpm *DBProviderMock) Update(usr *User, usernameToUpdate string) error {
	if dpm.UsersToUpdate == nil {
		dpm.UsersToUpdate = []*User{}
	}

	dpm.UsersToUpdate = append(dpm.UsersToUpdate, usr)

	return dpm.ErrorToGive
}

type FileManagerMock struct {
	UsersToRead  []*User
	WrittenUsers []*User
	ErrorToGive  error
}

func (fmm *FileManagerMock) ReadUsersFromFile() ([]*User, error) {
	return fmm.UsersToRead, fmm.ErrorToGive
}

func (fmm *FileManagerMock) SaveUsersToFile(users []*User) error {
	fmm.WrittenUsers = users
	return fmm.ErrorToGive
}

func TestGetUsersFromDB(t *testing.T) {
	givenUsers := []*User{
		{
			Username: "one",
			Password: "two",
			Groups:   []string{"group1"},
		},
	}
	db := &DBProviderMock{
		UsersToGive: givenUsers,
	}

	service := APIService{
		ProviderType: ProviderFromDB,
		DB:           db,
	}

	actualUsers, err := service.GetAll()

	require.NoError(t, err)
	assert.Equal(t, givenUsers, actualUsers)

	db = &DBProviderMock{
		UsersToGive: givenUsers,
		ErrorToGive: errors.New("some db error"),
	}

	service = APIService{
		ProviderType: ProviderFromDB,
		DB:           db,
	}

	_, err = service.GetAll()
	require.EqualError(t, err, "some db error")
}

func TestGetUsersFromFile(t *testing.T) {
	givenUsers := []*User{
		{
			Username: "user1",
			Password: "pass1",
			Groups:   []string{"group1", "group2"},
		},
	}

	service := APIService{
		ProviderType: ProviderFromFile,
		FileProvider: &FileManagerMock{
			UsersToRead: givenUsers,
		},
	}

	actualUsers, err := service.GetAll()

	require.NoError(t, err)
	assert.Equal(t, givenUsers, actualUsers)

	service = APIService{
		ProviderType: ProviderFromFile,
		FileProvider: &FileManagerMock{
			ErrorToGive: errors.New("some file error"),
		},
	}

	_, err = service.GetAll()
	require.EqualError(t, err, "some file error")
}

func TestUnsupportedUserProvider(t *testing.T) {
	service := APIService{
		ProviderType: ProviderFromStaticPassword,
	}

	_, err := service.GetAll()
	require.EqualError(t, err, fmt.Sprintf("unknown user data provider type: %d", ProviderFromStaticPassword))
}
