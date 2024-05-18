package models

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
)

type Setting struct {
	Verbose bool    `json:"verbose" validate:"required"`
	Files   []Files `json:"files" validate:"required"`
	Backup  Backup  `json:"backup" validate:"required"`
	Yandex  Yandex  `json:"yandex" validate:"required"`
}

type Files struct {
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type Backup struct {
	Dir       string   `json:"dir" validate:"required"`
	Retention int      `json:"count" validate:"required"`
	Expired   Duration `json:"expired" validate:"required"`
}

type Yandex struct {
	Timeout   Duration `json:"timeout" validate:"required"`
	Token     string   `json:"token" validate:"required"`
	Dir       string   `json:"dir" validate:"required"`
	Extension bool     `json:"extension" validate:"required"`
}

type IError struct {
	Field string
	Tag   string
	Value string
}

func (s *Setting) Validate() error {
	v := validator.New()
	err := v.Struct(s)
	var errs []IError

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var el IError
			el.Field = err.Field()
			el.Tag = err.Tag()
			el.Value = err.Param()
			errs = append(errs, el)
		}

		message, err := json.Marshal(errs)

		if err != nil {
			return err
		}

		return errors.New(string(message))
	} else {
		return nil
	}
}
