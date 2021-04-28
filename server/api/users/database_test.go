package users

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserDatabase(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = prepareTables(db)
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE `invalid_users` (username TEXT PRIMARY KEY, pass TEXT)")
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE `invalid_groups` (username TEXT, other TEXT)")
	require.NoError(t, err)

	testCases := []struct {
		Name          string
		UsersTable    string
		GroupsTable   string
		ExpectedError string
	}{
		{
			Name:          "invalid users tables",
			UsersTable:    "non_existent_users",
			GroupsTable:   "groups",
			ExpectedError: "no such table: non_existent_users",
		}, {
			Name:          "invalid groups tables",
			UsersTable:    "users",
			GroupsTable:   "non_existent_groups",
			ExpectedError: "no such table: non_existent_groups",
		}, {
			Name:          "invalid users columns",
			UsersTable:    "invalid_users",
			GroupsTable:   "groups",
			ExpectedError: "no such column: password",
		}, {
			Name:          "invalid groups columns",
			UsersTable:    "users",
			GroupsTable:   "invalid_groups",
			ExpectedError: "no such column: group",
		}, {
			Name:        "valid tables",
			UsersTable:  "users",
			GroupsTable: "groups",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := NewUserDatabase(db, tc.UsersTable, tc.GroupsTable)
			if tc.ExpectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
			}
		})
	}
}

func TestGetByUsername(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = prepareTables(db)
	require.NoError(t, err)

	err = prepareDummyData(db)
	require.NoError(t, err)

	d, err := NewUserDatabase(db, "users", "groups")
	require.NoError(t, err)

	testCases := []struct {
		Name         string
		Username     string
		ExpectedUser *User
	}{
		{
			Name:         "non existent user",
			Username:     "user99",
			ExpectedUser: nil,
		}, {
			Name:     "user without groups",
			Username: "user1",
			ExpectedUser: &User{
				Username: "user1",
				Password: "pass1",
			},
		}, {
			Name:     "user with one group",
			Username: "user2",
			ExpectedUser: &User{
				Username: "user2",
				Password: "pass2",
				Groups:   []string{"group1"},
			},
		}, {
			Name:     "user with multiple groups",
			Username: "user3",
			ExpectedUser: &User{
				Username: "user3",
				Password: "pass3",
				Groups:   []string{"group1", "group2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			u, err := d.GetByUsername(tc.Username)
			require.NoError(t, err)

			assert.Equal(t, tc.ExpectedUser, u)
		})
	}

}

func TestGetAll(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = prepareTables(db)
	require.NoError(t, err)

	err = prepareDummyData(db)
	require.NoError(t, err)

	d, err := NewUserDatabase(db, "users", "groups")
	require.NoError(t, err)

	actualUsers, err := d.GetAll()
	require.NoError(t, err)

	expectedUsers := []*User{
		{
			Username: "user1",
			Password: "pass1",
			Groups:   nil,
		},
		{
			Username: "user2",
			Password: "pass2",
			Groups: []string{
				"group1",
			},
		},
		{
			Username: "user3",
			Password: "pass3",
			Groups: []string{
				"group1",
				"group2",
			},
		},
	}
	assert.Equal(t, expectedUsers, actualUsers)
}

func TestAdd(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = prepareTables(db)
	require.NoError(t, err)

	d, err := NewUserDatabase(db, "users", "groups")
	require.NoError(t, err)

	givenUser := &User{
		Username: "login1",
		Password: "pass1",
		Groups: []string{
			"group1",
			"group2",
		},
	}

	err = d.Add(givenUser)
	require.NoError(t, err)

	actualUser := User{}

	err = d.db.Get(&actualUser, fmt.Sprintf("SELECT username, password FROM `%s`", d.usersTableName))
	require.NoError(t, err)
	assert.Equal(t, givenUser.Username, actualUser.Username)
	assert.Equal(t, givenUser.Password, actualUser.Password)

	type group struct {
		Username string `db:"username"`
		Group    string `db:"group"`
	}

	actualGroups := []group{}
	err = d.db.Select(&actualGroups, fmt.Sprintf("SELECT `username`, `group` FROM `%s`", d.groupsTableName))
	require.NoError(t, err)

	assert.Equal(
		t,
		[]group{
			{
				Username: "login1",
				Group: "group1",
			},
			{
				Username: "login1",
				Group: "group2",
			},
		},
		actualGroups,
	)
}

func prepareTables(db *sqlx.DB) error {
	_, err := db.Exec("CREATE TABLE `users` (username TEXT PRIMARY KEY, password TEXT)")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE `groups` (username TEXT, `group` TEXT)")
	if err != nil {
		return err
	}

	return nil
}

func prepareDummyData(db *sqlx.DB) error {
	_, err := db.Exec("INSERT INTO `users` (username, password) VALUES (\"user1\", \"pass1\")")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO `users` (username, password) VALUES (\"user2\", \"pass2\")")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO `users` (username, password) VALUES (\"user3\", \"pass3\")")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO `groups` (username, `group`) VALUES (\"user2\", \"group1\")")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO `groups` (username, `group`) VALUES (\"user3\", \"group1\")")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO `groups` (username, `group`) VALUES (\"user3\", \"group2\")")
	if err != nil {
		return err
	}

	return nil
}
