package persistence

import (
	"time"
)

type (
	persistentMock struct {
	}
)

//func NewPersistentMock() Persistent {
//	return &persistentMock{}
//}

func (persistent *persistentMock) CheckPassword(login, password string) (bool, error) {
	return true, nil
}

func (persistent *persistentMock) GetUserAuthParams(login string) (User, error) {
	user := user{
		IdUser:   1,
		Username: "1",
		Role:     "user",
	}
	return &user, nil
}
func (persistent *persistentMock) AddUser(login string, password string) (User, error) {
	user := user{
		IdUser:   1,
		Username: "1",
		Role:     "user",
	}
	return &user, nil
}
func (persistent *persistentMock) DeleteUser(id int64) bool {
	return true
}

//func (persistent *persistentMock) GetUserProfile(id int64) (UserProfile, error) {
//	//profile := UserProfileImpl{
//	//	IdUser:     1,
//	//	Name:       sql.NullString{String: "Василий", Valid: true},
//	//	SecondName: sql.NullString{String: "Пупкин", Valid: true},
//	//	Sex:        sql.NullString{String: "male", Valid: true},
//	//	Height:     sql.NullInt64{Int64: 180, Valid: true},
//	//	Weight:     sql.NullInt64{Int64: 80, Valid: true},
//	//	Email:      "vasya@gmai.com",
//	//	Location:   sql.NullString{String: "Петроградская", Valid: true},
//	//	About:      sql.NullString{String: "Я люблю бегать!", Valid: true},
//	//	IdLevel:    sql.NullInt64{Int64: 1, Valid: true},
//	//}
//	return nil, nil
//}
//func (persistent *persistentMock) UpdateUserProfile(profile persistence.per) bool {
//	return true
//}

func (persistent *persistentMock) GetLevel(id int64) (Level, error) {
	lvl := level{
		id,
		1,
		"Новичок",
	}
	return &lvl, nil
}

func (persistent *persistentMock) AddSession(token Token) bool {
	return true
}

func (persistent *persistentMock) GetSession(tknId string) (tokenResp Token, err error) {
	tkn := token{
		Token:     tknId,
		IdUser:    1,
		LoginDate: time.Now(),
	}
	return &tkn, err
}

func (persistent *persistentMock) RemoveSession(tknId string) error {
	return nil
}

func (persistent *persistentMock) UpdateSession(tkn Token) error {
	return nil
}
func (persistent *persistentMock) GetRole(id int64) string {
	return ""
}
