package validator

import "github.com/go-playground/validator/v10"

type NotificationValidator struct {
	validate *validator.Validate
}

func New() *NotificationValidator {
	return &NotificationValidator{validate: validator.New()}
}
func (v *NotificationValidator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}
