package persistence

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type (
	// Duration lets us convert between a bigint in Postgres and time.Duration
	// in Go
	Duration time.Duration

	user struct {
		IdUser   int64 `gorm:"primaryKey"`
		Username string
		Role     string
	}

	userAuthInfo struct {
		IdUser   int64
		Login    string
		Password string
	}

	User interface {
		GetId() int64
		GetRole() string
		GetUsername() string
	}

	UserProfileImpl struct {
		// Id of a user
		IdUser int64 `json:"id_user" example:"709786"`
		// First name of a user
		Name string `json:"name" example:"Андрей"`
		// Second name of a user
		SecondName string `json:"second_name" example:"Попов"`
		// Sex of a user
		Sex string `json:"sex" example:"male"`
		// Height of a user
		Height int64 `json:"height" example:"180"`
		// Weight of a user
		Weight int64 `json:"weight" example:"80"`
		// User's e-mail
		Email string `json:"email" example:"andrey@gmail.com"`
		// Level id. Can be 1 - junior, 2 - middle, 3 - senior
		IdLevel int64 `json:"id_level" example:"1"`
		// Preferred location (metro station)
		Location string `json:"location" example:"Петроградская"`
		// Date of birth of a user
		DateOfBirth time.Time `json:"date_of_birth,omitempty" example:"2000-01-01T00:00:00Z"`
		// Description of a user
		About string `json:"about" example:"Я люблю бегать"`
	} // @name UserProfile

	FilteredUserProfileImpl struct {
		// Id of a user
		IdUser int64 `json:"id_user" example:"709786"`
		// First name of a user
		Name string `json:"name" example:"Андрей"`
		// Second name of a user
		SecondName string `json:"second_name" example:"Попов"`
		// Sex of a user
		Sex string `json:"sex" example:"male"`
		// Height of a user
		Height int64 `json:"height" example:"180"`
		// Weight of a user
		Weight int64 `json:"weight" example:"80"`
		// Level id. Can be 1 - junior, 2 - middle, 3 - senior
		IdLevel int64 `json:"id_level" example:"1"`
		// Preferred location (metro station)
		Location string `json:"location" example:"Петроградская"`
		// Date of birth of a user
		Age int `json:"age" example:"20"`
		// Description of a user
		About  string `json:"about" example:"Я люблю бегать"`
		Sports string `json:"-" swag:"-"`
		// Kinds of sports of a user
		PersonSports []string `json:"person_sports" gorm:"-" example:"волейбол"`
	} // @name FilteredProfile

	token struct {
		Token     string
		IdUser    int64
		Username  string
		Role      string
		LoginDate time.Time
		Expires   time.Time
	}

	Token interface {
		GetUsername() string
		GetId() string
		GetRole() string
		GetUserId() int64
		GetLoginDate() time.Time
		GetExpirationTime() time.Time
	}

	level struct {
		IdLevel     int64
		Level       int
		Description string
	}

	Level interface {
		GetId() int64
		GetLevel() int
		GetDescription() string
	}

	sport struct {
		IdSport   int64 `gorm:"primaryKey"`
		SportType string
	}

	personSports struct {
		IdUser  int64
		IdSport int64
	}

	Filter interface {
		BuildMapAndQuery() (string, map[string]interface{})
	}

	groupTraining struct {
		IdTraining       int64     `json:"id_training,omitempty" gorm:"primaryKey"`
		Owner            int64     `json:"owner" gorm:"-"`
		MeetDate         time.Time `json:"meet_date"`
		Duration         Duration  `json:"-"`
		TrainingDuration string    `json:"training_duration" gorm:"-"`
		Location         string    `json:"location"`
		Sport            string    `json:"sport"`
		IdSport          int64     `json:"-"`
		IdLevel          int64     `json:"id_level,omitempty"`
		Comment          string    `json:"comment,omitempty"`
		Fee              int64     `json:"fee,omitempty"`
		ParticipantsIds  []int64   `json:"participants_ids,omitempty" gorm:"-"`
		Kind             string    `json:"kind"`
	}

	memberTraining struct {
		IdUser        int64
		IdTraining    int64
		TrainingOwner bool
	}

	message struct {
		IdMes     int64     `gorm:"primaryKey"`
		IdTo      int64     `json:"id_to"`
		IdFrom    int64     `json:"id_from"`
		Content   string    `json:"content"`
		CreatedAt time.Time `json:"created_at"`
	}

	dialog struct {
		IdTo      int64     `json:"id_to"`
		IdFrom    int64     `json:"id_from"`
		Seen      bool      `json:"seen"`
		CreatedAt time.Time `json:"created_at" gorm:"created_at"`
		Type      string    `json:"type"`
		Status    string    `json:"status,omitempty"`
	}

	PersistentObject interface {
		Serialize() []byte
	}
)

func (gt *groupTraining) Serialize() []byte {
	res, _ := json.Marshal(gt)
	return res
}

func (u *user) GetId() int64 {
	return u.IdUser
}
func (u *user) GetRole() string {
	return u.Role
}

func (u *user) GetUsername() string {
	return u.Username
}

func (ua *userAuthInfo) GetId() int64 {
	return ua.IdUser
}

func (ua *userAuthInfo) GetLogin() string {
	return ua.Login
}

func (token *token) GetId() string {
	return token.Token
}
func (token *token) GetUserId() int64 {
	return token.IdUser
}
func (token *token) GetLoginDate() time.Time {
	return token.LoginDate
}
func (token *token) GetExpirationTime() time.Time {
	return token.Expires
}
func (token *token) GetRole() string {
	return token.Role
}

func (token *token) GetUsername() string {
	return token.Username
}

func (lvl *level) GetId() int64 {
	return lvl.IdLevel
}
func (lvl *level) GetLevel() int {
	return lvl.Level
}
func (lvl *level) GetDescription() string {
	return lvl.Description
}

func (prf *UserProfileImpl) Serialize() []byte {
	prof, _ := json.Marshal(prf)
	return prof
}

// Value converts Duration to a primitive value ready to written to a database.
func (d Duration) Value() (driver.Value, error) {
	return driver.Value(int64(d)), nil
}

// Scan reads a Duration value from database driver type.
func (d *Duration) Scan(raw interface{}) error {
	switch v := raw.(type) {
	case int64:
		*d = Duration(v)
	case nil:
		*d = Duration(0)
	//case string:
	//	dur, err:=time.ParseDuration()
	//	if err!=nil{
	//		log.Errorf("hello from string: %s", v)
	//		return fmt.Errorf("cannot sql.Scan() strfmt.Duration from: %#v", v)
	//	}
	//	*d = Duration(dur)
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Duration from: %#v", v)
	}
	return nil
}

func (dialog *dialog) Serialize() []byte {
	res, _ := json.Marshal(dialog)
	return res
}

func (msg *message) Serialize() []byte {
	res, _ := json.Marshal(msg)
	return res
}
