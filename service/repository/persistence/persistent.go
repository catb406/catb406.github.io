package persistence

import (
	"encoding/json"
	"errors"
	"gorm.io/gorm/clause"
	"strings"
	"time"

	//"context"
	//"database/sql"
	//"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
	//"gorm.io/gorm/schema"
	//"strconv"
)

type (
	Persistent interface {
		GetRole(id int64) string
		GetUserAuthParams(login string) (User, error)
		AddUser(login string, password string) (User, error)
		DeleteUser(id int64) bool
		GetUserProfile(id int64) (PersistentObject, error)
		UpdateUserProfile(profile PersistentObject, sports ...string) bool
		AddSession(token Token) bool
		GetSession(id string) (Token, error)
		RemoveSession(id string) error
		CheckPassword(login, password string) (bool, error)
		GetUserSport(id int64) []string
		GetFilteredProfiles(filter Filter) []FilteredUserProfileImpl
		GetGroupTrainings(filter Filter) []PersistentObject
		GetGroupTraining(id int64) PersistentObject
		AddGroupTraining(training PersistentObject) (PersistentObject, error)
		UpdateGroupTraining(training PersistentObject) (PersistentObject, error)
		DeleteGroupTraining(id int64) error
		GetUserTrainings(idUser int64) ([]PersistentObject, error)
		AddMessage(msg PersistentObject) error
		AddRequest(request PersistentObject) error
		UpdateRequest(request PersistentObject) (PersistentObject, error)
		GetDialogs(idUser int64) ([]PersistentObject, error)
		GetMessages(idUsers []int64, t time.Time) ([]PersistentObject, error)
		//GetLevel(id int64) (Level, error)
	}

	persistent struct {
		db *gorm.DB
		//db *sql.DB
	}
)

func NewPersistent(dbConnection *gorm.DB) Persistent {
	return &persistent{
		db: dbConnection,
	}
}

func (persistent *persistent) GetRole(id int64) string {
	role := ""
	res := persistent.db.Table(`users`).Select(`role`).Where(`id_user=?`, id).Find(&role)
	if res.Error == gorm.ErrRecordNotFound {
		log.Error(fmt.Errorf("no user with id %d", id))
		return ""
	} else if res.Error != nil {
		log.Error(res.Error)
		return ""
	}
	return role
}

func (persistent *persistent) CheckPassword(login, password string) (bool, error) {
	match := struct {
		Pswmatch bool
	}{}
	res := persistent.db.Table(`user_auth_info`).Select(`(password = crypt(?, password)) AS pswmatch`, password).Where(`login=?`, login).Find(&match)
	if err := res.Error; err == gorm.ErrRecordNotFound {
		log.Errorf("no user with login %s", login)
		return false, fmt.Errorf("no user with login %s", login)
	} else if res.Error != nil {
		log.Error(res.Error)
		return false, errors.New("failed to check password")
	}

	return match.Pswmatch, nil
}

func (persistent *persistent) GetUserAuthParams(login string) (User, error) {
	params := user{
		Username: login,
	}
	res := persistent.db.Table(`users`).Select(`id_user, role`).Where(`username=?`, &params.Username).Find(&params)
	if res.Error == gorm.ErrRecordNotFound {
		log.Error(fmt.Errorf("no user with login %s", login))
		return nil, fmt.Errorf("no user with login %s", login)
	} else if res.Error != nil {
		log.Error(res.Error)
		return nil, errors.New("failed to get user auth params")
	}
	log.Info("user auth params ", params)
	return &params, nil
}
func (persistent *persistent) AddUser(login string, password string) (userParams User, err error) {
	log.Info("add user start")
	params := user{
		Username: login,
	}
	userParams = &params
	tx := persistent.db.Begin()
	defer func(err error) {
		if r := recover(); r != nil {
			log.Error("rolling back")
			tx.Rollback()
			err = errors.New("failed to create user")
		}
	}(err)

	if err := tx.Error; err != nil {
		log.Error(err)
		return nil, errors.New("problem with database")
	}

	res := tx.Table(`users`).Select(`username`).Create(&params)
	if err := res.Error; err != nil {
		tx.Rollback()
		log.Error(err)
		return nil, errors.New("failed to create user")
	}
	params.Role = "user"

	log.Info(params)

	res = tx.Table(`user_info`).Select(`id_user`).Create(&params)
	if err := res.Error; err != nil {
		tx.Rollback()
		log.Error(err)
		return nil, errors.New("failed to create user")
	}

	res = tx.Exec(`INSERT INTO user_auth_info  (id_user, login, password) VALUES (?, ?,  crypt(?, gen_salt('md5')));`, params.IdUser, params.Username, password)
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		tx.Rollback()
		log.Error(err)
		return nil, errors.New("failed to create user")
	}

	if err := tx.Commit().Error; err != nil {
		log.Error(err)
		return nil, errors.New("failed to create user")
	}
	return
}

func (persistent *persistent) DeleteUser(idUser int64) bool {
	tx := persistent.db.Begin()
	subQuery := tx.Table(`member_training`).Select(`id_training`).Where(`id_user=? AND training_owner=true`, idUser)
	res := tx.Table(`group_training`).Where(`id_training IN ?`, subQuery).Delete(&groupTraining{})
	if err := res.Error; err != nil {
		tx.Rollback()
		log.Error(err)
		return false
	}
	sqlStatement := `DELETE FROM users WHERE id_user = ?;`
	res = tx.Exec(sqlStatement, idUser)
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		tx.Rollback()
		log.Error(err)
		return false
	}
	tx.Commit()
	return true
}

func (persistent *persistent) GetUserProfile(id int64) (PersistentObject, error) {
	var profile UserProfileImpl
	profile.IdUser = id
	res := persistent.db.Table(`user_info`).Where(`id_user=?`, profile.IdUser).First(&profile)
	if err := res.Error; err != nil {
		log.Error(err)
		return nil, err
	}
	return &profile, nil
}

func (persistent *persistent) UpdateUserProfile(profile PersistentObject, sports ...string) bool {
	p := UserProfileImpl{}
	err := json.Unmarshal(profile.Serialize(), &p)
	if err != nil {
		log.Error(err)
		return false
	}
	tx := persistent.db.Begin()
	res := tx.Table(`user_info`).Where(`id_user=?`, p.IdUser).Updates(&p)
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		log.Error(err)
		tx.Rollback()
		return false
	}

	if len(sports) > 0 {
		var sportIds []int64

		var st []sport
		var sp []personSports
		for _, s := range sports {
			st = append(st, sport{SportType: s})
		}

		res = tx.Table(`sports`).Where(`sport_type IN ?`, sports).Find(&st)
		if res.Error != nil {
			tx.Rollback()
			log.Error("cannot find sports ", sports)
			return false
		}

		for _, s := range st {
			sp = append(sp, personSports{
				IdUser:  p.IdUser,
				IdSport: s.IdSport,
			})
		}
		res = tx.Table(`person_sports`).Clauses(clause.OnConflict{DoNothing: true}).Create(&sp)
		if res.Error != nil {
			tx.Rollback()
			log.Error("error occurred while updating person's sports: ", res.Error)
			return false
		}

		for _, s := range sp {
			sportIds = append(sportIds, s.IdSport)
		}
		res = tx.Table(`person_sports`).Where(`id_user=? AND id_sport NOT IN ?`, p.IdUser, sportIds).Delete(&personSports{})
		if res.Error != nil {
			tx.Rollback()
			log.Error("error occurred while updating person's sports: ", res.Error)
			return false
		}
	} else {
		res = tx.Table(`person_sports`).Where(`id_user=?`, p.IdUser).Delete(&sport{})
		if res.Error != nil {
			tx.Rollback()
			log.Error("error occurred while updating person's sports: ", res.Error)
			return false
		}
	}

	tx.Commit()
	return true
}

func (persistent *persistent) AddSession(token Token) bool {
	sqlStatement := `INSERT INTO sessions (id_user, login_time, token, expires) VALUES (?, ?, ?, ?);`
	res := persistent.db.Exec(sqlStatement, token.GetUserId(), token.GetLoginDate(), token.GetId(), token.GetExpirationTime())
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		log.Error(err)
		return false
	}
	return true
}

func (persistent *persistent) GetSession(tknId string) (tokenResp Token, err error) {
	tkn := token{
		Token: tknId,
	}
	res := persistent.db.Table(`sessions`).Select(` login_time, token, expires, sessions.id_user, username, role`).Joins(`join users u on u.id_user = sessions.id_user`).Where(&tkn).Take(&tkn)
	if err = res.Error; err != nil {
		log.Error(err)
		return nil, err
	}
	return &tkn, err
}

func (persistent *persistent) RemoveSession(tknId string) error {
	res := persistent.db.Table(`sessions`).Where("token=?", tknId).Delete(&token{})
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		log.Error(err)
		return err
	}
	log.Info("session ", tknId, " deleted")
	return nil
}

func (persistent *persistent) GetFilteredProfiles(filter Filter) []FilteredUserProfileImpl {
	var filtered []FilteredUserProfileImpl
	var res *gorm.DB
	sub := persistent.db.Table(`user_info`).Select(`user_info.id_user,user_info.id_level,
user_info.name,user_info.second_name,user_info.sex,user_info.height,
user_info.weight,user_info.location, user_info.about,
extract(year from age(now(), user_info.date_of_birth)) as age, array_agg(sports.sport_type) as sports`).Joins(`left join person_sports on 
person_sports.id_user=user_info.id_user left join sports on person_sports.id_sport=sports.id_sport`).Group(`user_info.id_user`)
	q, m := filter.BuildMapAndQuery()
	if len(m) != 0 {
		res = persistent.db.Table(`(?) as u`, sub).Where(q, m).Find(&filtered)
	} else {
		res = persistent.db.Table(`(?) as u`, sub).Find(&filtered)
	}
	if res.Error == gorm.ErrRecordNotFound {
		log.Error("no user with such parameters")
		return nil
	} else if res.Error != nil {
		log.Error(res.Error)
		return nil
	}
	var sports []string
	s, ok := m["sports"]
	if ok {
		sports = s.([]string)
	}
	var result []FilteredUserProfileImpl
	for i := range filtered {
		if filtered[i].Sports != "{NULL}" {
			withoutBrackets := strings.Trim(filtered[i].Sports, "{}")
			filtered[i].PersonSports = strings.Split(withoutBrackets, ",")
		}
		match := false
	LOOP:
		for _, sp := range sports {
			for _, personSport := range filtered[i].PersonSports {
				if sp == personSport {
					match = true
					break LOOP
				}
			}
		}
		if match || sports == nil {
			result = append(result, filtered[i])
		}
	}
	return result
}

func (persistent *persistent) GetUserSport(id int64) []string {
	var sportType []string
	res := persistent.db.Table(`person_sports`).Select(`s.sport_type`).Joins(`join sports s on s.id_sport = person_sports.id_sport`).Where(`id_user=?`, id).Find(&sportType)
	if res.Error == gorm.ErrRecordNotFound {
		log.Error(fmt.Errorf("no user with id %d", id))
		return nil
	} else if res.Error != nil {
		log.Error(res.Error)
		return nil
	}
	return sportType
}
func (persistent *persistent) GetGroupTraining(idTraining int64) PersistentObject {
	result := groupTraining{}
	res := persistent.db.Table(`group_training`).Select(`group_training.kind, group_training.id_training, group_training.location, group_training.meet_date, group_training.duration, s.sport_type as sport, group_training.id_level, group_training.comment, group_training.fee`).Joins(`JOIN sports s on group_training.id_sport = s.id_sport`).Where(`id_training=?`, idTraining).Find(&result)
	if res.Error == gorm.ErrRecordNotFound {
		log.Error("no training with id ", idTraining)
		return nil
	} else if res.Error != nil {
		log.Error(res.Error)
		return nil
	}
	var mts []memberTraining
	res = persistent.db.Table(`member_training`).Where(`id_training=?`, result.IdTraining).Find(&mts)
	if res.Error == gorm.ErrRecordNotFound {
		log.Errorf("no members in training with id %s", result.IdTraining)
		return nil
	} else if res.Error != nil {
		log.Error(res.Error)
		return nil
	}
	for _, mt := range mts {
		result.ParticipantsIds = append(result.ParticipantsIds, mt.IdUser)
		if mt.TrainingOwner {
			result.Owner = mt.IdUser
		}
	}
	result.TrainingDuration = time.Duration(result.Duration).String()
	return &result
}
func (persistent *persistent) GetGroupTrainings(filter Filter) []PersistentObject {
	var result []PersistentObject
	var filtered []groupTraining
	var res *gorm.DB
	subQuery := persistent.db.Table(`group_training`).Select(`group_training.kind, group_training.id_training, group_training.location, group_training.meet_date, group_training.duration, s.sport_type as sport, group_training.id_level, group_training.comment, group_training.fee`).Joins(`JOIN sports s on group_training.id_sport = s.id_sport`).Where(`meet_date > now() AND kind = 'group'`)
	q, m := filter.BuildMapAndQuery()
	if len(m) > 0 {
		res = persistent.db.Table(`(?) as t`, subQuery).Where(q, m).Order(`meet_date DESC`).Find(&filtered)
		if res.Error == gorm.ErrRecordNotFound {
			log.Error("no training with such parameters")
			return nil
		} else if res.Error != nil {
			log.Error(res.Error)
			return nil
		}
	} else {
		res = persistent.db.Table(`(?) as t`, subQuery).Order(`meet_date DESC`).Find(&filtered)
		if res.Error == gorm.ErrRecordNotFound {
			log.Error("no training with such parameters")
			return nil
		} else if res.Error != nil {
			log.Error(res.Error)
			return nil
		}
	}
	for i, t := range filtered {
		var mts []memberTraining
		res = persistent.db.Table(`member_training`).Where(`id_training=?`, t.IdTraining).Find(&mts)
		if res.Error == gorm.ErrRecordNotFound {
			log.Errorf("no members in training with id %s", t.IdTraining)
			return nil
		} else if res.Error != nil {
			log.Error(res.Error)
			return nil
		}
		for _, mt := range mts {
			filtered[i].ParticipantsIds = append(filtered[i].ParticipantsIds, mt.IdUser)
			if mt.TrainingOwner {
				filtered[i].Owner = mt.IdUser
			}
		}
		filtered[i].TrainingDuration = time.Duration(filtered[i].Duration).String()
		result = append(result, &filtered[i])
	}
	return result
}

func (persistent *persistent) AddGroupTraining(training PersistentObject) (PersistentObject, error) {
	var gt groupTraining
	err := json.Unmarshal(training.Serialize(), &gt)
	if err != nil {
		log.Errorf("failed to unmarshal group training: %s", err)
		return nil, fmt.Errorf("failed to unmarshal group training: %s", err)
	}
	d, err := time.ParseDuration(gt.TrainingDuration)
	if err == nil {
		gt.Duration = Duration(d)
	}
	s, err := persistent.getSport(gt.Sport)
	if err != nil {
		return nil, err
	}
	gt.IdSport = s.IdSport
	tx := persistent.db.Begin()
	res := tx.Table(`group_training`).Select(`meet_date`, `location`, `id_sport`, `id_level`, `duration`, `comment`, `fee`, `kind`).Create(&gt)
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		tx.Rollback()
		log.Error(err)
		return nil, fmt.Errorf("failed to add group training: %s", err)
	}
	mt := memberTraining{
		IdUser:        gt.Owner,
		IdTraining:    gt.IdTraining,
		TrainingOwner: true,
	}
	res = tx.Table(`member_training`).Create(&mt)
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		tx.Rollback()
		log.Error(err)
		return nil, fmt.Errorf("failed to add group training: %s", err)
	}
	gt.ParticipantsIds = []int64{gt.Owner}
	tx.Commit()
	gt.TrainingDuration = time.Duration(gt.Duration).String()
	return &gt, nil
}

func (persistent *persistent) UpdateGroupTraining(training PersistentObject) (PersistentObject, error) {
	var gt groupTraining
	err := json.Unmarshal(training.Serialize(), &gt)
	if err != nil {
		log.Errorf("failed to unmarshal group training: %s", err)
		return nil, fmt.Errorf("failed to unmarshal group training: %s", err)
	}
	d, err := time.ParseDuration(gt.TrainingDuration)
	if err == nil {
		gt.Duration = Duration(d)
	}
	s, err := persistent.getSport(gt.Sport)
	if err != nil {
		return nil, err
	}
	gt.IdSport = s.IdSport
	tx := persistent.db.Begin()
	res := tx.Table(`group_training`).Select(`meet_date`, `location`, `id_sport`, `id_level`, `duration`, `comment`, `fee`, `kind`).Updates(&gt)
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		tx.Rollback()
		log.Error(err)
		return nil, errors.New("nothing changed")
	}
	var mt []memberTraining
	trainingOwner := false
	for _, u := range gt.ParticipantsIds {
		if u == gt.Owner {
			trainingOwner = true
			mt = append(mt, memberTraining{
				IdUser:        u,
				IdTraining:    gt.IdTraining,
				TrainingOwner: true,
			})
		} else {
			mt = append(mt, memberTraining{
				IdUser:        u,
				IdTraining:    gt.IdTraining,
				TrainingOwner: false,
			})
		}
	}
	if gt.Owner == 0 || !trainingOwner {
		tx.Rollback()
		log.Error("no owner provided or trying to delete owner from participants")
		return nil, errors.New("no owner provided or trying to delete owner from participants")
	}
	res = tx.Table(`member_training`).Clauses(clause.OnConflict{DoNothing: true}).Create(&mt)
	if err := res.Error; err != nil {
		tx.Rollback()
		log.Error(err)
		return nil, fmt.Errorf("failed to add group training: %s", err)
	}
	res = tx.Table(`member_training`).Where(`id_training=? AND id_user NOT IN ? AND training_owner!=?`, gt.IdTraining, gt.ParticipantsIds, trainingOwner).Delete(&memberTraining{})
	if err = res.Error; err != nil {
		tx.Rollback()
		log.Error(err)
		return nil, fmt.Errorf("failed to add group training: %s", err)
	}
	tx.Commit()
	gt.TrainingDuration = time.Duration(gt.Duration).String()
	return &gt, nil
}

func (persistent *persistent) DeleteGroupTraining(idTraining int64) error {
	res := persistent.db.Table(`group_training`).Where(`id_training=?`, idTraining).Delete(&groupTraining{})
	if err := res.Error; err != nil || res.RowsAffected == 0 {
		log.Error(err)
		return err
	}
	return nil
}

func (persistent *persistent) getSport(s string) (sport, error) {
	resSport := sport{SportType: s}
	res := persistent.db.Table(`sports`).Where(`sport_type=?`, resSport.SportType).First(&resSport)
	if res.Error == gorm.ErrRecordNotFound {
		log.Errorf("no sport with name %s", s)
		return sport{}, fmt.Errorf("no sport with name %s", s)
	} else if res.Error != nil {
		log.Errorf("failed to get sport: %s", res.Error)
		return sport{}, fmt.Errorf("failed to get sport: %s", res.Error)
	}
	return resSport, nil
}

func (persistent *persistent) GetUserTrainings(idUser int64) ([]PersistentObject, error) {
	var gt []groupTraining
	var result []PersistentObject
	subQuery := persistent.db.Table(`group_training`).Select(`group_training.id_training, meet_date, location, sport_type as sport, id_level, fee, kind, duration, comment`).
		Joins(`join member_training mt on group_training.id_training = mt.id_training join sports s on group_training.id_sport = s.id_sport`).Where(`id_user=?`, idUser)
	res := persistent.db.Table(`(?) as gt`, subQuery).Select(`member_training.id_training, meet_date, location, sport, id_level, fee, kind, duration, comment, member_training.id_user as owner `).
		Joins(`join member_training on gt.id_training=member_training.id_training`).Where(`training_owner=true`).Find(&gt)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	} else if res.Error != nil {
		log.Errorf("failed to get trainings: %s", res.Error)
		return nil, fmt.Errorf("failed to get trainings: %s", res.Error)
	}

	for _, t := range gt {
		var mts []memberTraining
		res = persistent.db.Table(`member_training`).Where(`id_training=?`, t.IdTraining).Find(&mts)
		if res.Error == gorm.ErrRecordNotFound {
			log.Errorf("no members in training with id %s", t.IdTraining)
			return nil, nil
		} else if err := res.Error; err != nil {
			log.Error(err)
			return nil, err
		}
		for _, mt := range mts {
			t.ParticipantsIds = append(t.ParticipantsIds, mt.IdUser)
		}

		t.TrainingDuration = time.Duration(t.Duration).String()
		result = append(result, &t)
	}
	return result, nil
}

func (persistent *persistent) AddMessage(msg PersistentObject) error {
	var m message
	err := json.Unmarshal(msg.Serialize(), &m)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("failed to unmarshal message: %s", err.Error())
	}
	res := persistent.db.Table(`messages`).Create(&m)
	if res.Error != nil || res.RowsAffected == 0 {
		err = fmt.Errorf("failed to add message to db: err")
		log.Error(err)
		return err
	}
	return nil
}

func (persistent *persistent) GetDialogs(idUser int64) ([]PersistentObject, error) {
	var dialogs []dialog
	log.Info("id_user", idUser)
	res := persistent.db.Table(`relationships`).Where(`(id_to=? OR id_from=?) AND ((status='declined' AND seen=false) OR status!='declined')`, idUser, idUser).Order(`created_at DESC`).Find(&dialogs)
	if res.Error != nil {
		err := fmt.Errorf("failed to get dialogs: %s", res.Error)
		log.Error(err)
		return nil, err
	}
	log.Info(dialogs)
	var result = make([]PersistentObject, len(dialogs))
	for i := range dialogs {
		result[i] = &dialogs[i]
	}
	return result, nil
}

func (persistent *persistent) AddRequest(request PersistentObject) error {
	var req dialog
	err := json.Unmarshal(request.Serialize(), &req)
	if err != nil {
		return fmt.Errorf("failed to unmarshal request: %s", err.Error())
	}
	req.Status = "request"
	req.Seen = false
	res := persistent.db.Table(`relationships`).Create(&req)
	if res.Error != nil || res.RowsAffected == 0 {
		err = fmt.Errorf("failed to add request: %s", res.Error)
		log.Error(err)
		return err
	}
	return nil
}

func (persistent *persistent) UpdateRequest(request PersistentObject) (PersistentObject, error) {
	var req dialog
	err := json.Unmarshal(request.Serialize(), &req)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %s", err.Error())
	}
	res := persistent.db.Table(`relationships`).Where(`id_to=? AND id_from=?`, &req.IdTo, &req.IdFrom).Updates(&req)
	if res.Error != nil || res.RowsAffected == 0 {
		err = fmt.Errorf("failed to update request: %s", res.Error)
		log.Error(err)
		return nil, err
	}
	res = persistent.db.Table(`relationships`).Where(`id_to=? AND id_from=?`, &req.IdTo, &req.IdFrom).Find(&req)
	return &req, nil
}

func (persistent *persistent) GetMessages(idUsers []int64, t time.Time) ([]PersistentObject, error) {
	var msg []message
	res := persistent.db.Table(`messages`).Where(`id_from IN ? AND id_to IN ? AND created_at<=?`, idUsers, idUsers, t).Order(`created_at DESC`).Limit(20).Find(&msg)
	if res.Error != nil {
		err := fmt.Errorf("failed to get messaged: %s", res.Error)
		return nil, err
	}
	var result = make([]PersistentObject, len(msg))
	for i := range msg {
		result[i] = &msg[i]
	}
	return result, nil
}
