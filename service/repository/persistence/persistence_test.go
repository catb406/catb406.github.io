package persistence

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	username        = "test"
	password        = "123"
	idUser    int64 = 11111
	userRole        = "user"
	mockToken       = token{
		Token:     "YHT2HFCRKALV7ZMQRMBFH6434T6PCZWD7X73AOJVRDCS3ERR",
		IdUser:    idUser,
		Username:  username,
		Role:      userRole,
		LoginDate: time.Unix(1637603397, 0),
		Expires:   time.Unix(1640184760, 0),
	}
	mockFilteredProfile = FilteredUserProfileImpl{
		IdUser: idUser, Name: "name", SecondName: "secondname", Sex: "male", Height: 150, Weight: 50, IdLevel: 1, Location: "Avtovo", Age: 20, About: "about", Sports: "{football}",
		PersonSports: mockSports,
	}
	mockGroupTraining = groupTraining{
		IdTraining:       1,
		Owner:            idUser,
		MeetDate:         time.Unix(1640184760, 0),
		Duration:         3600000000000,
		TrainingDuration: "1h0m0s",
		Location:         "Avtovo",
		Sport:            "football",
		IdLevel:          1,
		Comment:          "some comment",
		Fee:              500,
		ParticipantsIds:  []int64{idUser},
		Kind:             "group",
	}
	mockUser         = &user{IdUser: idUser, Username: username, Role: userRole}
	mockBirthDate, _ = time.Parse("2006-04-02", "2000-01-01")
	mockSports       = []string{"football"}
	mockSportType    = sport{IdSport: 1, SportType: "football"}
	mockProfile      = &UserProfileImpl{IdUser: idUser, Name: "name", SecondName: "secondname", Sex: "male", Height: 150, Weight: 50, Email: "email@example.com", IdLevel: 1, Location: "Avtovo", DateOfBirth: mockBirthDate, About: "about"}
)

type (
	Suite struct {
		suite.Suite
		DB   *gorm.DB
		mock sqlmock.Sqlmock

		persistent Persistent
	}

	userFilterMock struct {
		Sex        string   `json:"sex,omitempty" example:"male"`
		WeightFrom int64    `json:"weight_from,omitempty" example:"80"`
		WeightTo   int64    `json:"weight_to,omitempty" example:"90"`
		IdLevel    []int64  `json:"id_level,omitempty" example:"1"`
		Sport      []string `json:"sport,omitempty" example:"волейбол"`
		Location   []string `json:"location,omitempty" example:"Петроградская"`
		AgeFrom    int64    `json:"age_from,omitempty" example:"20"`
		AgeTo      int64    `json:"age_to,omitempty" example:"40"`
	}

	groupTrainingMockFilter struct {
		Location []string `json:"location,omitempty"`
		Sport    []string `json:"sport,omitempty"`
		IdLevel  []int64  `json:"id_level,omitempty"`
	}
)

func (filter *userFilterMock) BuildMapAndQuery() (string, map[string]interface{}) {
	result := make(map[string]interface{})
	queryParts := make([]string, 0)
	if filter.AgeFrom != 0 {
		result["age_from"] = filter.AgeFrom
		queryParts = append(queryParts, `age >= @age_from`)
	}
	if filter.AgeTo != 0 {
		result["age_to"] = filter.AgeTo
		queryParts = append(queryParts, `age <= @age_to`)

	}
	if filter.WeightFrom != 0 {
		result["weight_from"] = filter.WeightFrom
		queryParts = append(queryParts, `weight >= @weight_from`)

	}
	if filter.WeightTo != 0 {
		result["weight_to"] = filter.WeightTo
		queryParts = append(queryParts, `weight <= @weight_to`)
	}
	if filter.IdLevel != nil {
		result["id_level"] = filter.IdLevel
		queryParts = append(queryParts, `id_level IN @id_level`)
	}
	if filter.Sport != nil {
		result["sports"] = filter.Sport
	}
	if filter.Location != nil {
		result["location"] = filter.Location
		queryParts = append(queryParts, `location IN @location`)
	}
	if filter.Sex != "" {
		result["sex"] = filter.Sex
		queryParts = append(queryParts, `sex = @sex`)
	}
	query := strings.Join(queryParts, " AND ")
	return query, result
}

func (gtf *groupTrainingMockFilter) BuildMapAndQuery() (string, map[string]interface{}) {
	queryParts := make([]string, 0)
	resMap := make(map[string]interface{})
	if gtf.Location != nil {
		resMap["location"] = gtf.Location
		queryParts = append(queryParts, `location IN @location`)
	}
	if gtf.Sport != nil {
		resMap["sports"] = gtf.Sport
		queryParts = append(queryParts, `sport IN @sports`)
	}
	if gtf.IdLevel != nil {
		resMap["id_level"] = gtf.IdLevel
		queryParts = append(queryParts, `id_level IN @id_level`)
	}
	query := strings.Join(queryParts, ` AND `)
	return query, resMap
}

func (s *Suite) SetupSuite() {
	var (
		psqlDb *sql.DB
		err    error
	)

	psqlDb, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	s.DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: psqlDb,
	}), &gorm.Config{})

	require.NoError(s.T(), err)

	s.DB.Debug()

	s.persistent = NewPersistent(s.DB)
}

func (s *Suite) TestGetUserAuthParams() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT id_user, role FROM "users" WHERE username=$1`)).WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id_user", "role"}).AddRow(idUser, userRole))

	u, err := s.persistent.GetUserAuthParams(username)
	require.NoError(s.T(), err)
	require.Equal(s.T(), mockUser, u)
}

func (s *Suite) TestCheckPassword() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT (password = crypt($1, password)) AS pswmatch FROM "user_auth_info" WHERE login=$2`)).WithArgs(password, username).
		WillReturnRows(sqlmock.NewRows([]string{"pswmatch"}).AddRow(true))

	res, err := s.persistent.CheckPassword(username, password)
	require.NoError(s.T(), err)
	require.Equal(s.T(), true, res)
}

func (s *Suite) TestGetRole() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT role FROM "users" WHERE id_user=$1`)).WithArgs(idUser).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(userRole))

	res := s.persistent.GetRole(idUser)
	require.Equal(s.T(), userRole, res)
}

func (s *Suite) TestAddUser() {
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("username") VALUES ($1) RETURNING "id_user`)).WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id_user"}).AddRow(idUser))
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user_info" ("id_user") VALUES ($1) RETURNING "id_user"`)).WithArgs(idUser).
		WillReturnRows(sqlmock.NewRows([]string{"id_user"}).AddRow(idUser))
	s.mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO user_auth_info  (id_user, login, password) VALUES ($1, $2,  crypt($3, gen_salt('md5')));`)).WithArgs(idUser, username, password).WillReturnResult(sqlmock.NewResult(idUser, 1))
	s.mock.ExpectCommit()

	res, err := s.persistent.AddUser(username, password)

	require.NoError(s.T(), err)
	require.Equal(s.T(), mockUser, res)
}

func (s *Suite) TestDeleteUser() {
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "group_training" WHERE id_training IN SELECT id_training FROM "member_training" WHERE id_user=$1 AND training_owner=true`)).WithArgs(idUser).
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id_user = $1;`)).WithArgs(idUser).
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.mock.ExpectCommit()
	res := s.persistent.DeleteUser(idUser)
	require.Equal(s.T(), true, res)
}

func (s *Suite) TestGetUserProfile() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_info" WHERE id_user=$1 ORDER BY "user_info"."id_user" LIMIT 1`)).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id_user", "name", "second_name", "sex", "height", "weight", "email", "id_level", "location", "date_of_birth", "about"}).AddRow(mockProfile.IdUser, mockProfile.Name,
			mockProfile.SecondName, mockProfile.Sex, mockProfile.Height, mockProfile.Weight, mockProfile.Email, mockProfile.IdLevel, mockProfile.Location, mockProfile.DateOfBirth, mockProfile.About))
	res, err := s.persistent.GetUserProfile(idUser)
	require.NoError(s.T(), err)
	require.Equal(s.T(), mockProfile, res)
}

func (s *Suite) TestUpdateUserProfile() {
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_info" SET "id_user"=$1,"name"=$2,"second_name"=$3,"sex"=$4,"height"=$5,"weight"=$6,"email"=$7,"id_level"=$8,"location"=$9,"date_of_birth"=$10,"about"=$11 WHERE id_user=$12`)).
		WithArgs(mockProfile.IdUser, mockProfile.Name,
			mockProfile.SecondName, mockProfile.Sex, mockProfile.Height, mockProfile.Weight, mockProfile.Email, mockProfile.IdLevel, mockProfile.Location, mockProfile.DateOfBirth, mockProfile.About, mockProfile.IdUser).
		WillReturnResult(sqlmock.NewResult(mockProfile.IdUser, 1))
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sports" WHERE sport_type IN ($1)`)).WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id_sport", "sport_type"}).AddRow(mockSportType.IdSport, mockSportType.SportType))
	s.mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "person_sports" ("id_user","id_sport") VALUES ($1,$2) ON CONFLICT DO NOTHING`)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(mockSportType.IdSport, 1))
	s.mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "person_sports" WHERE id_user=$1 AND id_sport NOT IN ($2)`)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()
	res := s.persistent.UpdateUserProfile(mockProfile, mockSports...)
	require.Equal(s.T(), true, res)
}

func (s *Suite) TestAddSession() {
	s.mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO sessions (id_user, login_time, token, expires) VALUES ($1, $2, $3, $4)`)).WithArgs(mockToken.IdUser, mockToken.LoginDate, mockToken.Token, mockToken.Expires).
		WillReturnResult(sqlmock.NewResult(1, 1))
	res := s.persistent.AddSession(&mockToken)
	require.Equal(s.T(), true, res)
}

func (s *Suite) TestGetSession() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT login_time, token, expires, sessions.id_user, username, role FROM "sessions" join users u on u.id_user = sessions.id_user WHERE "sessions"."token" = $1 LIMIT 1`)).WithArgs(mockToken.Token).
		WillReturnRows(sqlmock.NewRows([]string{"id_user", "login_date", "token", "expires", "username", "role"}).AddRow(mockToken.IdUser, mockToken.LoginDate, mockToken.Token, mockToken.Expires, mockToken.Username, mockToken.Role))

	res, err := s.persistent.GetSession(mockToken.Token)
	require.NoError(s.T(), err)
	require.Equal(s.T(), &mockToken, res)
}

func (s *Suite) TestRemoveSessions() {
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sessions" WHERE token=$1`)).WithArgs(mockToken.Token).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()
	err := s.persistent.RemoveSession(mockToken.Token)
	require.NoError(s.T(), err)
}

func (s *Suite) TestGetFilteredProfiles() {
	filter := &userFilterMock{
		Sex:   "female",
		Sport: []string{"football", "tennis"},
	}
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM (SELECT user_info.id_user,user_info.id_level,
user_info.name,user_info.second_name,user_info.sex,user_info.height,
user_info.weight,user_info.location, user_info.about,
extract(year from age(now(), user_info.date_of_birth)) as age, array_agg(sports.sport_type) as sports FROM "user_info" left join person_sports on 
person_sports.id_user=user_info.id_user left join sports on person_sports.id_sport=sports.id_sport GROUP BY "user_info"."id_user") as u WHERE sex = $1`)).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id_user", "id_level", "name", "second_name", "sex",
			"height", "weight", "location", "about", "age", "sports"}).AddRow(mockFilteredProfile.IdUser, mockFilteredProfile.IdLevel, mockFilteredProfile.Name, mockFilteredProfile.SecondName, mockFilteredProfile.Sex,
			mockFilteredProfile.Height, mockFilteredProfile.Weight, mockFilteredProfile.Location, mockFilteredProfile.About, mockFilteredProfile.Age, mockFilteredProfile.Sports))

	res := s.persistent.GetFilteredProfiles(filter)
	require.Equal(s.T(), []FilteredUserProfileImpl{mockFilteredProfile}, res)
}

func (s *Suite) TestGetUserSport() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT s.sport_type FROM "person_sports" join sports s on s.id_sport = person_sports.id_sport WHERE id_user=$1`)).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"sport_type"}).AddRow("football"))

	res := s.persistent.GetUserSport(idUser)
	require.Equal(s.T(), mockSports, res)
}

func (s *Suite) TestGetGroupTrainings() {
	filter := groupTrainingMockFilter{
		Location: []string{"Avtovo"},
		Sport:    []string{"football"},
	}
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM (SELECT group_training.kind, group_training.id_training, group_training.location, group_training.meet_date, group_training.duration, s.sport_type as sport, group_training.id_level, group_training.comment, group_training.fee FROM "group_training" JOIN sports s on group_training.id_sport = s.id_sport WHERE meet_date > now() AND kind = 'group') as t WHERE location IN ($1) AND sport IN ($2) ORDER BY meet_date DESC`)).
		WithArgs(mockProfile.Location, mockFilteredProfile.PersonSports[0]).
		WillReturnRows(sqlmock.NewRows([]string{"kind", "id_training", "location", "meet_date", "duration", "sport", "id_level", "comment", "fee"}).
			AddRow(mockGroupTraining.Kind, mockGroupTraining.IdTraining, mockGroupTraining.Location, mockGroupTraining.MeetDate, mockGroupTraining.Duration, mockGroupTraining.Sport, mockGroupTraining.IdLevel, mockGroupTraining.Comment, mockGroupTraining.Fee))
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_training" WHERE id_training=$1`)).WithArgs(mockGroupTraining.IdTraining).
		WillReturnRows(sqlmock.NewRows([]string{"id_user", "id_training", "training_owner"}).AddRow(idUser, mockGroupTraining.IdTraining, true))
	res := s.persistent.GetGroupTrainings(&filter)
	require.Equal(s.T(), []PersistentObject{&mockGroupTraining}, res)
}

func (s *Suite) TestAddGroupTraining() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sports" WHERE sport_type=$1 ORDER BY "sports"."id_sport" LIMIT 1`)).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id_sport", "sport_type"}).AddRow(0, "football"))
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "group_training" ("meet_date","duration","location","id_sport","id_level","comment","fee","kind") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id_training"`)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id_training"}).AddRow(mockGroupTraining.IdTraining))
	s.mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_training" ("id_user","id_training","training_owner") VALUES ($1,$2,$3)`)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(idUser, 1))
	res, err := s.persistent.AddGroupTraining(&mockGroupTraining)
	require.NoError(s.T(), err)
	require.Equal(s.T(), &mockGroupTraining, res)
}

func (s *Suite) TestUpdateGroupTraining() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sports" WHERE sport_type=$1 ORDER BY "sports"."id_sport" LIMIT 1`)).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id_sport", "sport_type"}).AddRow(0, "football"))
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "group_training" SET "meet_date"=$1,"duration"=$2,"location"=$3,"id_sport"=$4,"id_level"=$5,"comment"=$6,"fee"=$7,"kind"=$8 WHERE "id_training" = $9`)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(idUser, 1))
	s.mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_training" ("id_user","id_training","training_owner") VALUES ($1,$2,$3)`)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(idUser, 1))
	s.mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "member_training" WHERE id_training=$1 AND id_user NOT IN ($2) AND training_owner!=$3`)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.mock.ExpectCommit()
	res, err := s.persistent.UpdateGroupTraining(&mockGroupTraining)
	require.NoError(s.T(), err)
	require.Equal(s.T(), &mockGroupTraining, res)
}

func (s *Suite) TestDeleteGroupTraining() {
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "group_training" WHERE id_training=$1`)).WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()
	err := s.persistent.DeleteGroupTraining(mockGroupTraining.IdTraining)
	require.NoError(s.T(), err)
}

func (s *Suite) TestGetUserTrainings() {
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT member_training.id_training, meet_date, location, sport, id_level, fee, kind, duration, comment, member_training.id_user as owner  
FROM (SELECT group_training.id_training, meet_date, location, sport_type as sport, id_level, fee, kind, duration, comment FROM "group_training" 
join member_training mt on group_training.id_training = mt.id_training join sports s on group_training.id_sport = s.id_sport WHERE id_user=$1) 
as gt join member_training on gt.id_training=member_training.id_training WHERE training_owner=true`)).WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id_training",
		"meet_date", "location", "sport", "id_level", "fee", "kind", "duration", "comment", "owner"}).AddRow(mockGroupTraining.IdTraining, mockGroupTraining.MeetDate, mockGroupTraining.Location, mockGroupTraining.Sport,
		mockGroupTraining.IdLevel, mockGroupTraining.Fee, mockGroupTraining.Kind, mockGroupTraining.Duration, mockGroupTraining.Comment, idUser))
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_training" WHERE id_training=$1`)).WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id_user", "id_training", "training_owner"}).AddRow(idUser, mockGroupTraining.IdTraining, true))

	trainings, err := s.persistent.GetUserTrainings(idUser)
	require.NoError(s.T(), err)
	require.Equal(s.T(), []PersistentObject{&mockGroupTraining}, trainings)
}

func (s *Suite) AfterTest(_, _ string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func TestPersistent(t *testing.T) {
	suite.Run(t, new(Suite))
}
