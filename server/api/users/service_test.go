package users

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type DBProviderMock struct {
	UsersToGive         []*User
	UsersToAdd          []*User
	UsersToUpdate       []*User
	ErrorToGiveOnRead   error
	ErrorToGiveOnWrite  error
	ErrorToGiveOnDelete error
	UsernameToUpdate    string
	UsernameToDelete    string
}

func (dpm *DBProviderMock) GetAll() ([]*User, error) {
	return dpm.UsersToGive, dpm.ErrorToGiveOnRead
}

func (dpm *DBProviderMock) GetByUsername(username string) (*User, error) {
	var usr *User
	for i := range dpm.UsersToGive {
		if dpm.UsersToGive[i].Username == username {
			usr = dpm.UsersToGive[i]
		}
	}

	return usr, dpm.ErrorToGiveOnRead
}

func (dpm *DBProviderMock) Add(usr *User) error {
	if dpm.UsersToAdd == nil {
		dpm.UsersToAdd = []*User{}
	}

	dpm.UsersToAdd = append(dpm.UsersToAdd, usr)

	return dpm.ErrorToGiveOnWrite
}

func (dpm *DBProviderMock) Update(usr *User, usernameToUpdate string) error {
	if dpm.UsersToUpdate == nil {
		dpm.UsersToUpdate = []*User{}
	}

	dpm.UsersToUpdate = append(dpm.UsersToUpdate, usr)
	dpm.UsernameToUpdate = usernameToUpdate

	return dpm.ErrorToGiveOnWrite
}

func (dpm *DBProviderMock) Delete(usernameToDelete string) error {
	dpm.UsernameToDelete = usernameToDelete
	return dpm.ErrorToGiveOnDelete
}

type FileManagerMock struct {
	UsersToRead        []*User
	WrittenUsers       []*User
	ErrorToGiveOnRead  error
	ErrorToGiveOnWrite error
}

func (fmm *FileManagerMock) ReadUsersFromFile() ([]*User, error) {
	return fmm.UsersToRead, fmm.ErrorToGiveOnRead
}

func (fmm *FileManagerMock) SaveUsersToFile(users []*User) error {
	fmm.WrittenUsers = users
	return fmm.ErrorToGiveOnWrite
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
		UsersToGive:       givenUsers,
		ErrorToGiveOnRead: errors.New("some db error"),
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
			ErrorToGiveOnRead: errors.New("some file error"),
		},
	}

	_, err = service.GetAll()
	require.EqualError(t, err, "some file error")
}

func TestAddUserToFile(t *testing.T) {
	givenUser := &User{
		Username: "user1",
		Password: "pass1",
		Groups:   []string{"group1", "group2"},
	}

	usersFileManager := &FileManagerMock{}
	service := APIService{
		ProviderType: ProviderFromFile,
		FileProvider: usersFileManager,
	}

	err := service.Change(givenUser, "")
	require.NoError(t, err)

	require.Len(t, usersFileManager.WrittenUsers, 1)
	assert.Equal(t, givenUser, usersFileManager.WrittenUsers[0])

	usersFileManager = &FileManagerMock{
		ErrorToGiveOnRead: errors.New("some read error"),
	}
	service = APIService{
		ProviderType: ProviderFromFile,
		FileProvider: usersFileManager,
	}
	err = service.Change(givenUser, "")
	require.EqualError(t, err, "some read error")

	usersFileManager = &FileManagerMock{
		ErrorToGiveOnWrite: errors.New("some write error"),
	}
	service = APIService{
		ProviderType: ProviderFromFile,
		FileProvider: usersFileManager,
	}
	err = service.Change(givenUser, "")
	require.EqualError(t, err, "some write error")
}

func TestAddUserIfItExists(t *testing.T) {
	givenUser := &User{
		Username: "user1",
		Password: "pass1",
	}

	usersFileManager := &FileManagerMock{
		UsersToRead: []*User{
			{
				Username: "user1",
				Password: "pass1",
			},
			{
				Username: "user2",
				Password: "pass2",
			},
		},
	}
	service := APIService{
		ProviderType: ProviderFromFile,
		FileProvider: usersFileManager,
	}

	err := service.Change(givenUser, "")
	require.EqualError(t, err, "Another user with this username already exists")
	require.Len(t, usersFileManager.WrittenUsers, 0)
}

func TestUnsupportedUserProvider(t *testing.T) {
	service := APIService{
		ProviderType: ProviderFromStaticPassword,
	}

	_, err := service.GetAll()
	require.EqualError(t, err, fmt.Sprintf("unknown user data provider type: %d", ProviderFromStaticPassword))

	userToUpdate := &User{
		Username: "user_one",
		Password: "pass_one",
	}
	err = service.Change(userToUpdate, "")
	require.EqualError(t, err, fmt.Sprintf("unknown user data provider type: %d", ProviderFromStaticPassword))

	err = service.Delete("some")
	require.EqualError(t, err, fmt.Sprintf("unknown user data provider type: %d", ProviderFromStaticPassword))
}

func TestValidate(t *testing.T) {
	service := APIService{
		ProviderType: ProviderFromFile,
		FileProvider: &FileManagerMock{
			UsersToRead: []*User{},
		},
	}

	testCases := []struct {
		name             string
		expectedError    string
		userKeyToProvide string
		user             *User
	}{
		{
			user:             &User{},
			expectedError:    "nothing to change",
			name:             "empty user on update",
			userKeyToProvide: "some",
		},
		{
			user:          &User{},
			expectedError: "username is required, password is required",
			name:          "empty user on create",
		},
		{
			user: &User{
				Password: "123",
			},
			expectedError: "username is required",
			name:          "no username provided",
		},
		{
			user: &User{
				Username: "someuser",
			},
			expectedError: "password is required",
			name:          "no password provided",
		},
		{
			user: &User{
				Username: "user123",
			},
			expectedError:    "nothing to change",
			name:             "nothing to change for the same username",
			userKeyToProvide: "user123",
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			err := service.Change(testCase.user, testCase.userKeyToProvide)
			require.EqualError(t, err, testCase.expectedError)
		})
	}
}

func TestUpdateUserInFile(t *testing.T) {
	userToUpdate := &User{
		Username: "user_one",
		Password: "pass_one",
		Groups:   []string{"group_one", "group_two"},
	}

	usersFileManager := &FileManagerMock{
		UsersToRead: []*User{
			{
				Username: "user2",
			},
			{
				Username: "user1",
				Password: "pass1",
				Groups:   []string{"group1", "group2"},
			},
		},
	}
	service := APIService{
		ProviderType: ProviderFromFile,
		FileProvider: usersFileManager,
	}

	err := service.Change(userToUpdate, "user1")
	require.NoError(t, err)

	require.Len(t, usersFileManager.WrittenUsers, 2)
	assert.Equal(t, usersFileManager.UsersToRead, usersFileManager.WrittenUsers)

	userToUpdate = &User{
		Username: "unknown_user",
		Password: "222",
	}
	err = service.Change(userToUpdate, "unknown_user")
	assert.EqualError(t, err, "cannot find user by username 'unknown_user'")

	service = APIService{
		ProviderType: ProviderFromFile,
		FileProvider: &FileManagerMock{
			ErrorToGiveOnWrite: errors.New("failed to write to file"),
			UsersToRead: []*User{
				{
					Username: "user2",
				},
			},
		},
	}

	userToUpdate = &User{
		Username: "user2",
		Password: "3342",
	}
	err = service.Change(userToUpdate, "user2")
	require.EqualError(t, err, "failed to write to file")
}

func TestAddUserToDB(t *testing.T) {
	givenUser := &User{
		Username: "user13",
		Password: "pass13",
	}

	dbProvider := &DBProviderMock{}
	service := APIService{
		ProviderType: ProviderFromDB,
		DB:           dbProvider,
	}

	err := service.Change(givenUser, "")
	require.NoError(t, err)

	require.Len(t, dbProvider.UsersToAdd, 1)
	assert.Equal(t, givenUser, dbProvider.UsersToAdd[0])
	require.Len(t, dbProvider.UsersToUpdate, 0)

	dbProvider = &DBProviderMock{
		ErrorToGiveOnRead: errors.New("some read error"),
	}
	service = APIService{
		ProviderType: ProviderFromDB,
		DB:           dbProvider,
	}
	err = service.Change(givenUser, "")
	require.EqualError(t, err, "some read error")

	dbProvider = &DBProviderMock{
		ErrorToGiveOnWrite: errors.New("some write error"),
	}
	service = APIService{
		ProviderType: ProviderFromDB,
		DB:           dbProvider,
	}
	err = service.Change(givenUser, "")
	require.EqualError(t, err, "some write error")
}

func TestAddUserToDBIfItExists(t *testing.T) {
	userToUpdate := &User{
		Username: "user1",
		Password: "pass1",
	}

	dbProvider := &DBProviderMock{
		UsersToGive: []*User{
			{
				Username: "user1",
				Password: "pass1",
			},
			{
				Username: "user2",
				Password: "pass2",
			},
		},
	}

	service := APIService{
		ProviderType: ProviderFromDB,
		DB:           dbProvider,
	}

	err := service.Change(userToUpdate, "")
	require.EqualError(t, err, "Another user with this username already exists")
	require.Len(t, dbProvider.UsersToAdd, 0)
	require.Len(t, dbProvider.UsersToUpdate, 0)
}

func TestUpdateUserToDBIfItExists(t *testing.T) {
	givenUser := &User{
		Username: "user1",
		Password: "pass1",
	}

	dbProvider := &DBProviderMock{
		UsersToGive: []*User{
			{
				Username: "user1",
				Password: "pass1",
			},
			{
				Username: "user2",
				Password: "pass2",
			},
		},
	}

	service := APIService{
		ProviderType: ProviderFromDB,
		DB:           dbProvider,
	}

	err := service.Change(givenUser, "user2")
	require.EqualError(t, err, "Another user with this username already exists")
	require.Len(t, dbProvider.UsersToAdd, 0)
	require.Len(t, dbProvider.UsersToUpdate, 0)
}

func TestUpdateUserInDB(t *testing.T) {
	userToUpdate := &User{
		Username: "user_one",
		Password: "pass_one",
		Groups:   []string{"group_one", "group_two"},
	}

	dbProvider := &DBProviderMock{
		UsersToGive: []*User{
			{
				Username: "user2",
			},
			{
				Username: "user1",
				Password: "pass1",
				Groups:   []string{"group1", "group2"},
			},
		},
	}
	service := APIService{
		ProviderType: ProviderFromDB,
		DB:           dbProvider,
	}

	err := service.Change(userToUpdate, "user1")
	require.NoError(t, err)

	require.Len(t, dbProvider.UsersToUpdate, 1)
	assert.Equal(t, userToUpdate, dbProvider.UsersToUpdate[0])
	assert.Equal(t, "user1", dbProvider.UsernameToUpdate)

	service = APIService{
		ProviderType: ProviderFromDB,
		DB: &DBProviderMock{
			ErrorToGiveOnWrite: errors.New("failed to write to DB"),
			UsersToGive: []*User{
				{
					Username: "user2",
				},
			},
		},
	}

	userToUpdate = &User{
		Username: "user2",
		Password: "3342",
	}
	err = service.Change(userToUpdate, "user2")
	require.EqualError(t, err, "failed to write to DB")
}

func TestDeleteUserFromDB(t *testing.T) {
	dbProvider := &DBProviderMock{}

	service := APIService{
		ProviderType: ProviderFromDB,
		DB:           dbProvider,
	}

	err := service.Delete("user2")
	require.NoError(t, err)
	assert.Equal(t, "user2", dbProvider.UsernameToDelete)

	dbProvider = &DBProviderMock{
		ErrorToGiveOnDelete: errors.New("failed to delete from db"),
	}

	service = APIService{
		ProviderType: ProviderFromDB,
		DB:           dbProvider,
	}

	err = service.Delete("user2")
	require.EqualError(t, err, "failed to delete from db")
}

func TestDeleteUserFromFile(t *testing.T) {
	usersFileManager := &FileManagerMock{
		UsersToRead: []*User{
			{
				Username: "user2",
			},
			{
				Username: "user1",
				Password: "pass1",
				Groups:   []string{"group1", "group2"},
			},
		},
	}

	service := APIService{
		ProviderType: ProviderFromFile,
		FileProvider: usersFileManager,
	}

	err := service.Delete("user2")
	require.NoError(t, err)
	expectedUsers := []*User{
		{
			Username: "user1",
			Password: "pass1",
			Groups:   []string{"group1", "group2"},
		},
	}
	assert.Equal(t, expectedUsers, usersFileManager.WrittenUsers)

	err = service.Delete("unknown_user")
	require.EqualError(t, err, "unknown user 'unknown_user'")

	usersFileManager = &FileManagerMock{
		ErrorToGiveOnRead: errors.New("failed to read users from file"),
	}
	service = APIService{
		ProviderType: ProviderFromFile,
		FileProvider: usersFileManager,
	}
	err = service.Delete("user2")
	require.EqualError(t, err, "failed to read users from file")

	usersFileManager = &FileManagerMock{
		UsersToRead: []*User{
			{
				Username: "user3",
			},
		},
		ErrorToGiveOnWrite: errors.New("failed to write users to file"),
	}
	service = APIService{
		ProviderType: ProviderFromFile,
		FileProvider: usersFileManager,
	}
	err = service.Delete("user3")
	require.EqualError(t, err, "failed to write users to file")
}
