package models

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type Databases struct {
	Paths []string `json:"paths" validate:"required"`
}

type Backup struct {
	Dir     string   `json:"dir" validate:"required"`
	Count   int      `json:"count" validate:"required"`
	Expired Duration `json:"expired" validate:"required"`
}

type Yandex struct {
	Timeout Duration `json:"timeout" validate:"required"`
	Token   string   `json:"token" validate:"required"`
	Dir     string   `json:"dir" validate:"required"`
}

type Settings struct {
	Verbose   bool      `json:"verbose" validate:"required"`
	Databases Databases `json:"databases" validate:"required"`
	Backup    Backup    `json:"backup" validate:"required"`
	Yandex    Yandex    `json:"yandex" validate:"required"`
}

type IError struct {
	Field string
	Tag   string
	Value string
}

func (s *Settings) Validate() error {
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
