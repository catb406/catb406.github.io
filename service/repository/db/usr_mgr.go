package db

import (
	"SB/catb406.github.io/service/repository/persistence"
	"encoding/json"
	"errors"
	"github.com/labstack/gommon/log"
)

type (
	userManager struct {
		persistent persistence.Persistent
	}

	UserManager interface {
		Authenticate(login, password string) (persistence.User, error)
		AddUser(login, password string) (persistence.User, error)
		DeleteUser(id int64) error
		GetUserProfile(id int64) (UserProfile, error)
		UpdateUserProfile(profile persistence.PersistentObject, sports ...string) error
		GetRole(id int64) string
		GetProfiles(filter UserProfileFilterParams) []persistence.FilteredUserProfileImpl
		GetUserTrainings(idUser int64) ([]persistence.PersistentObject, error)
	}
)

func NewDbManager(persistent persistence.Persistent) UserManager {
	return &userManager{
		persistent: persistent,
	}
}

func (usrMgr *userManager) Authenticate(login, password string) (persistence.User, error) {
	access, err := usrMgr.persistent.CheckPassword(login, password)
	user := userImpl{
		Username: login,
	}
	if err != nil {
		return nil, err
	}
	if !access {
		return nil, errors.New("invalid password")
	}
	authParams, err := usrMgr.persistent.GetUserAuthParams(login)
	if err != nil {
		return nil, err
	}
	user.IdUser = authParams.GetId()
	user.Username = authParams.GetUsername()
	user.Role = authParams.GetRole()
	return &user, nil
}

func (usrMgr *userManager) AddUser(login, password string) (persistence.User, error) {
	authParams, err := usrMgr.persistent.AddUser(login, password)
	if err != nil {
		return nil, err
	}
	user := userImpl{
		IdUser:   authParams.GetId(),
		Username: authParams.GetUsername(),
		Role:     authParams.GetRole(),
	}
	return &user, nil
}

func (usrMgr *userManager) DeleteUser(id int64) error {
	deleted := usrMgr.persistent.DeleteUser(id)
	if !deleted {
		return errors.New("failed to delete user")
	}
	return nil
}

func (usrMgr *userManager) GetUserProfile(id int64) (UserProfile, error) {
	profile, err := usrMgr.persistent.GetUserProfile(id)
	if err != nil {
		return UserProfile{}, err
	}
	prof := UserProfile{}
	err = json.Unmarshal(profile.Serialize(), &prof)
	if err != nil {
		log.Error("failed to unmarshal profile: ", err)
		return UserProfile{}, err
	}
	prof.Sport = usrMgr.persistent.GetUserSport(id)
	return prof, nil
}

func (usrMgr *userManager) UpdateUserProfile(profile persistence.PersistentObject, sports ...string) error {
	log.Info(profile)
	updated := usrMgr.persistent.UpdateUserProfile(profile, sports...)
	if !updated {
		return errors.New("failed to update profile")
	}
	return nil
}

func (usrMgr *userManager) GetRole(id int64) string {
	return usrMgr.persistent.GetRole(id)
}

func (usrMgr *userManager) GetProfiles(filter UserProfileFilterParams) []persistence.FilteredUserProfileImpl {
	profiles := usrMgr.persistent.GetFilteredProfiles(&filter)
	return profiles
}

func (usrMgr *userManager) GetUserTrainings(idUser int64) ([]persistence.PersistentObject, error) {
	trainings, err := usrMgr.persistent.GetUserTrainings(idUser)
	if err != nil {
		return nil, err
	}
	return trainings, nil
}
