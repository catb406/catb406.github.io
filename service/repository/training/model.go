package training

import (
	"encoding/json"
	"strings"
	"time"
)

type (
	GroupTraining struct {
		IdTraining       int64         `json:"id_training,omitempty"`
		Owner            int64         `json:"owner"`
		MeetDate         time.Time     `json:"meet_date"`
		TrainingDuration string        `json:"training_duration"`
		Duration         time.Duration `json:"-"`
		Location         string        `json:"location"`
		Sport            string        `json:"sport"`
		IdLevel          int64         `json:"id_level,omitempty"`
		Comment          string        `json:"comment,omitempty"`
		Fee              int64         `json:"fee,omitempty"`
		ParticipantsIds  []int64       `json:"participants_ids"`
		Kind             string        `json:"kind"`
	}

	GroupTrainingFilter struct {
		Location []string `json:"location,omitempty"`
		Sport    []string `json:"sport,omitempty"`
		IdLevel  []int64  `json:"id_level,omitempty"`
	}
)

func (gt *GroupTraining) Serialize() []byte {
	res, _ := json.Marshal(gt)
	return res
}

func (gtf *GroupTrainingFilter) BuildMapAndQuery() (string, map[string]interface{}) {
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
