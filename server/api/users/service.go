package users

import (
	"context"
	"fmt"
	"github.com/cloudradar-monitoring/rport/server/api/errors"
	"net/http"
)

type ProviderType int

const (
	ProviderFromFile ProviderType = iota + 1
	ProviderFromStaticPassword
	ProviderFromDB
)

type APIService struct {
	ProviderType ProviderType
	FilePath     string
	AuthPath     string
	DB           *UserDatabase
}

func (as *APIService) GetAll(ctx context.Context) ([]*User, error) {
	if as.ProviderType == ProviderFromStaticPassword {
		return nil, errors.APIError{
			Code: http.StatusBadRequest,
			Message: "server runs on a static user-password pair, please use JSON file or database for user data",
		}
	}

	if as.ProviderType == ProviderFromFile {
		authUsers, err := GetUsersFromFile(as.FilePath)
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
