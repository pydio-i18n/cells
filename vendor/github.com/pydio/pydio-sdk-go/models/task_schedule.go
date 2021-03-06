// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// TaskSchedule task schedule
// swagger:model Task_schedule
type TaskSchedule struct {

	// schedule type
	ScheduleType string `json:"scheduleType,omitempty"`

	// schedule value
	ScheduleValue string `json:"scheduleValue,omitempty"`
}

// Validate validates this task schedule
func (m *TaskSchedule) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *TaskSchedule) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *TaskSchedule) UnmarshalBinary(b []byte) error {
	var res TaskSchedule
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
