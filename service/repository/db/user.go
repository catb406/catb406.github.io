package db

type (
	userImpl struct {
		Username string `json:"username"`
		IdUser   int64  `json:"id_user"`
		Role     string `json:"role"`
		//Profile  UserProfile `json:"profile"`
	}
)

//func CreateUser(username string, id int64) User {
//	user := userImpl{
//		IdUser:   id,
//		Username: username,
//	}
//	return &user
//}

func (user *userImpl) GetUsername() string {
	return user.Username
}

func (user *userImpl) GetId() int64 {
	return user.IdUser
}

func (user *userImpl) GetRole() string {
	return user.Role
}

//func (user *userImpl) GetProfile() map[string]string {
//	return user.Profile
//}
//
//func (user *userImpl) UpdateProfile(profile map[string]string) {
//	user.Profile = profile
//}
