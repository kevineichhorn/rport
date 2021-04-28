package users

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type DBProviderMock struct {
	UsersToGive []*User
	ErrorToGive error
}

func (dpm *DBProviderMock) GetAll() ([]*User, error) {
	return dpm.UsersToGive, dpm.ErrorToGive
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
		DB: db,
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
		DB: db,
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
		FileProvider: func() ([]*User, error) {
			return givenUsers, nil
		},
	}

	actualUsers, err := service.GetAll()

	require.NoError(t, err)
	assert.Equal(t, givenUsers, actualUsers)

	service = APIService{
		ProviderType: ProviderFromFile,
		FileProvider: func() ([]*User, error) {
			return givenUsers, errors.New("some file error")
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
