package users

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type UserDatabase struct {
	db              *sqlx.DB
	usersTableName  string
	groupsTableName string
}

func NewUserDatabase(DB *sqlx.DB, usersTableName, groupsTableName string) (*UserDatabase, error) {
	d := &UserDatabase{
		db:              DB,
		usersTableName:  usersTableName,
		groupsTableName: groupsTableName,
	}
	if err := d.checkDatabaseTables(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *UserDatabase) checkDatabaseTables() error {
	_, err := d.db.Exec(fmt.Sprintf("SELECT username, password FROM `%s` LIMIT 0", d.usersTableName))
	if err != nil {
		return err
	}
	_, err = d.db.Exec(fmt.Sprintf("SELECT username, `group` FROM `%s` LIMIT 0", d.groupsTableName))
	if err != nil {
		return err
	}
	return nil
}

func (d *UserDatabase) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := d.db.Get(user, fmt.Sprintf("SELECT username, password FROM `%s` WHERE username = ? LIMIT 1", d.usersTableName), username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	err = d.db.Select(&user.Groups, fmt.Sprintf("SELECT DISTINCT(`group`) FROM `%s` WHERE username = ?", d.groupsTableName), username)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return user, nil
}

func (d *UserDatabase) GetAll() ([]*User, error) {
	var usrs []*User
	err := d.db.Select(&usrs, fmt.Sprintf("SELECT username, password FROM `%s` ORDER BY username", d.usersTableName))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var groups []struct {
		Username string `db:"username"`
		Group    string `db:"group"`
	}
	err = d.db.Select(&groups, fmt.Sprintf("SELECT `username`, `group` FROM `%s` ORDER BY username", d.groupsTableName))
	if err != nil {
		if err == sql.ErrNoRows {
			return usrs, nil
		}
		return nil, err
	}
	for i := range groups {
		for y := range usrs {
			if usrs[y].Username == groups[i].Username {
				usrs[y].Groups = append(usrs[y].Groups, groups[i].Group)
			}
		}
	}

	return usrs, nil
}

func (d *UserDatabase) Add(usr *User) error {
	_, err := d.db.Exec(
		fmt.Sprintf("INSERT INTO `%s` (`username`, `password`) VALUES (?, ?)", d.usersTableName),
		usr.Username,
		usr.Password,
	)
	if err != nil {
		return err
	}

	for i := range usr.Groups {
		_, err := d.db.Exec(
			fmt.Sprintf("INSERT INTO `%s` (`username`, `group`) VALUES (?, ?)", d.groupsTableName),
			usr.Username,
			usr.Groups[i],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *UserDatabase) Update(usr *User, usernameToUpdate string) error {
	if usernameToUpdate == "" {
		return errors.New("cannot update user with empty username")
	}

	params := []interface{}{}
	statements := []string{}
	if usr.Password != "" {
		statements = append(statements, "`password` = ?")
		params = append(params, usr.Password)
	}

	if usr.Username != "" && usr.Username != usernameToUpdate {
		statements = append(statements, "`username` = ?")
		params = append(params, usr.Username)
	}

	if len(params) > 0 {
		q := fmt.Sprintf(
			"UPDATE `%s` SET %s WHERE username = ?",
			d.usersTableName,
			strings.Join(statements, ", "),
		)
		params = append(params, usernameToUpdate)
		_, err := d.db.Exec(q, params...)
		if err != nil {
			return err
		}
	}

	if usr.Username != "" && usernameToUpdate != usr.Username {
		_, err := d.db.Exec(
			fmt.Sprintf("UPDATE `%s` SET `username` = ? WHERE `username` = ?", d.groupsTableName),
			usr.Username,
			usernameToUpdate,
		)
		if err != nil {
			return err
		}
	}

	groupUserName := usernameToUpdate
	if usr.Username != "" {
		groupUserName = usr.Username
	}

	if len(usr.Groups) > 0 {
		_, err := d.db.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `username` = ?", d.groupsTableName), groupUserName)
		if err != nil {
			return err
		}
	}

	for i := range usr.Groups {
		group := usr.Groups[i]
		_, err := d.db.Exec(
			fmt.Sprintf("INSERT INTO `%s` (`username`, `group`) VALUES (?, ?)", d.groupsTableName),
			groupUserName,
			group,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *UserDatabase) Delete(usernameToDelete string) error {
	_, err := d.db.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `username` = ?", d.usersTableName), usernameToDelete)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `username` = ?", d.groupsTableName), usernameToDelete)
	if err != nil {
		return err
	}

	return nil
}
