package training

import (
	"SB/catb406.github.io/service/repository/persistence"
	"fmt"
)

type (
	trainingManager struct {
		persistent persistence.Persistent
	}

	TrainingManager interface {
		GetTrainings(filter GroupTrainingFilter) []persistence.PersistentObject
		AddTraining(training GroupTraining) (persistence.PersistentObject, error)
		UpdateTraining(training GroupTraining) (persistence.PersistentObject, error)
		GetTraining(id int64) persistence.PersistentObject
		DeleteTraining(id int64) error
	}
)

func NewTrainingManager(persistent persistence.Persistent) TrainingManager {
	return &trainingManager{
		persistent: persistent,
	}
}

func (tm *trainingManager) GetTrainings(filter GroupTrainingFilter) []persistence.PersistentObject {
	gt := tm.persistent.GetGroupTrainings(&filter)
	return gt
}

func (tm *trainingManager) GetTraining(id int64) persistence.PersistentObject {
	gt := tm.persistent.GetGroupTraining(id)
	return gt
}

func (tm *trainingManager) AddTraining(training GroupTraining) (persistence.PersistentObject, error) {
	t, err := tm.persistent.AddGroupTraining(&training)
	if err != nil {
		return nil, fmt.Errorf("failed to add training: %s", err)
	}
	return t, err
}

func (tm *trainingManager) UpdateTraining(training GroupTraining) (persistence.PersistentObject, error) {
	t, err := tm.persistent.UpdateGroupTraining(&training)
	if err != nil {
		return nil, err
	}
	return t, err
}
func (tm *trainingManager) DeleteTraining(id int64) error {
	err := tm.persistent.DeleteGroupTraining(id)
	if err != nil {
		return fmt.Errorf("failed to delete training: %s", err)
	}
	return nil
}
