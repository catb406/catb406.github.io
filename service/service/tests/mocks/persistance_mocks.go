package mocks

import (
	"github.com/DATA-DOG/go-sqlmock"
)

var username = "test"

func ExpectMockGetUserAuthParams(mock sqlmock.Sqlmock) {
	// 	res := persistent.db.Table(`users`).Select(`id_user, role`).Where(`username=?`, &params.Username).Take(&params)
	mock.ExpectQuery(`SELECT id_user, role FROM users WHERE username=$1`).WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id_user", "role"}).AddRow(00001, "user"))
}

func ExpectCheckPassword(mock sqlmock.Sqlmock) {
	// 	res := persistent.db.Table(`users`).Select(`id_user, role`).Where(`username=?`, &params.Username).Take(&params)
	mock.ExpectQuery(`SELECT (password = crypt(?, password)) AS pswmatch user_auth_info WHERE (login = $1)`).WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"pswmatch"}).AddRow(true))
}
