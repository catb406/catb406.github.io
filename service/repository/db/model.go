package db

import (
	"encoding/json"
	"strings"
	"time"
)

type (
	UserProfile struct {
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
		DateOfBirth time.Time `json:"date_of_birth" example:"2000-01-01T00:00:00Z"`
		// Description of a user
		About string `json:"about" example:"Я люблю бегать"`
		// User's kinds of sport
		Sport []string `json:"sport" example:"волейбол"`
	} // @name UserProfile

	UserProfileFilterParams struct {
		// Sex of a user  (if "", then not specified)
		Sex string `json:"sex,omitempty" example:"male"`
		// Minimum weight of a user  (if -1, then not specified)
		WeightFrom int64 `json:"weight_from,omitempty" example:"80"`
		// Maximum weight of a user (if -1, then not specified)
		WeightTo int64 `json:"weight_to,omitempty" example:"90"`
		// Level ids
		IdLevel []int64 `json:"id_level,omitempty" example:"1"`
		// Kinds of sport
		Sport []string `json:"sport,omitempty" example:"волейбол"`
		// Preferred location (metro station)
		Location []string `json:"location,omitempty" example:"Петроградская"`
		// Minimum age of a user  (if -1, then not specified)
		AgeFrom int64 `json:"age_from,omitempty" example:"20"`
		// Maximum age of a user  (if -1, then not specified)
		AgeTo int64 `json:"age_to,omitempty" example:"40"`
	} // @name UserProfileFilterParams
)

func (prf *UserProfile) GetDateOfBirth() time.Time {
	return prf.DateOfBirth
}

func (prf *UserProfile) Serialize() []byte {
	prof, _ := json.Marshal(prf)
	return prof
}

func (prf *UserProfile) GetId() int64 {
	return prf.IdUser
}
func (prf *UserProfile) GetName() string {
	return prf.Name
}
func (prf *UserProfile) GetSecondName() string {
	return prf.SecondName
}
func (prf *UserProfile) GetSex() string {
	return prf.Sex
}
func (prf *UserProfile) GetHeight() int64 {
	return prf.Height
}
func (prf *UserProfile) GetWeight() int64 {
	return prf.Weight
}
func (prf *UserProfile) GetEmail() string {
	return prf.Email
}
func (prf *UserProfile) GetIdLevel() int64 {
	return prf.IdLevel
}
func (prf *UserProfile) GetLocation() string {
	return prf.Location
}
func (prf *UserProfile) GetAbout() string {
	return prf.About
}

func (filter *UserProfileFilterParams) GetMinimumAge() (int64, bool) {
	if filter.AgeFrom > 0 {
		return filter.AgeFrom, true
	}
	return 0, false
}

func (filter *UserProfileFilterParams) GetMaximumAge() (int64, bool) {
	if filter.AgeTo > 0 {
		return filter.AgeTo, true
	}
	return 0, false
}
func (filter *UserProfileFilterParams) GetMinimumWeight() (int64, bool) {
	if filter.WeightFrom > 0 {
		return filter.WeightFrom, true
	}
	return 0, false
}
func (filter *UserProfileFilterParams) GetMaximumWeight() (int64, bool) {
	if filter.WeightTo > 0 {
		return filter.WeightTo, true
	}
	return 0, false
}
func (filter *UserProfileFilterParams) GetSex() string {
	return filter.Sex
}
func (filter *UserProfileFilterParams) GetLocations() []string {
	return filter.Location
}
func (filter *UserProfileFilterParams) GetSports() []string {
	return filter.Sport
}
func (filter *UserProfileFilterParams) GetLevels() []int64 {
	return filter.IdLevel
}

func (filter *UserProfileFilterParams) BuildMapAndQuery() (string, map[string]interface{}) {
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
