package api

import (
	"encoding/json"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type RestJob struct {
	Id        string    `json:"id,omitempty"`
	OwnerId   string    `json:"-"`
	InputId   string    `json:"input_id,omitempty"`
	OutputId  string    `json:"output_id,omitempty"`
	Status    string    `json:"status,omitempty"`
	Error     string    `json:"error,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Steps     []*Step   `json:"steps,omitempty"`
}

func (j *Job) MarshalJSON() ([]byte, error) {
	return json.Marshal(&RestJob{
		Id:        j.GetId(),
		InputId:   j.GetInputId(),
		OutputId:  j.GetOutputId(),
		Status:    j.GetStatus(),
		Error:     j.GetError(),
		StartTime: j.GetStartTime().AsTime(),
		EndTime:   j.GetEndTime().AsTime(),
	})
}

func (j *Job) UnmarshalJSON(data []byte) error {
	rj := &RestJob{}
	if err := json.Unmarshal(data, rj); err != nil {
		return err
	}

	j.Id = rj.Id
	j.OwnerId = rj.OwnerId
	j.InputId = rj.InputId
	j.OutputId = rj.OutputId
	j.Status = rj.Status
	j.StartTime = timestamppb.New(rj.StartTime)
	j.EndTime = timestamppb.New(rj.EndTime)

	return nil
}

type RestStep struct {
	TaskType string   `json:"task_type,omitempty"`
	Args     []string `json:"args,omitempty"`
}

func (s *Step) MarshalJSON() ([]byte, error) {
	return json.Marshal(&RestStep{
		TaskType: s.TaskType,
		Args:     s.Args,
	})
}

type RestStorage struct {
	Id         string    `json:"id,omitempty"`
	OwnerId    string    `json:"-"`
	Name       string    `json:"name,omitempty"`
	Status     string    `json:"status,omitempty"`
	LastUsed   time.Time `json:"last_used,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

func (s *Storage) MarshalJSON() ([]byte, error) {
	return json.Marshal(&RestStorage{
		Id:         s.GetId(),
		Name:       s.GetName(),
		Status:     s.GetStatus(),
		LastUsed:   s.GetLastUsed().AsTime(),
		CreateTime: s.GetCreateTime().AsTime(),
	})
}

func (s *Storage) UnmarshalJSON(data []byte) error {
	rs := &RestStorage{}
	if err := json.Unmarshal(data, rs); err != nil {
		return err
	}

	s.Id = rs.Id
	s.OwnerId = rs.OwnerId
	s.Name = rs.Name
	s.Status = rs.Status
	s.LastUsed = timestamppb.New(rs.LastUsed)
	s.CreateTime = timestamppb.New(rs.CreateTime)

	return nil
}
