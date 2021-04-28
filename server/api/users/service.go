package users

import (
	"fmt"
)

type ProviderType int

const (
	ProviderFromFile ProviderType = iota + 1
	ProviderFromStaticPassword
	ProviderFromDB
)

type DatabaseProvider interface {
	GetAll() ([]*User, error)
}

type APIService struct {
	ProviderType ProviderType
	FileProvider func() ([]*User, error)
	DB           DatabaseProvider
}

func (as *APIService) GetAll() ([]*User, error) {
	if as.ProviderType == ProviderFromFile {
		authUsers, err := as.FileProvider()
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
