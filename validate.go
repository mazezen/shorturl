package main

import (
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/validator/v10"
)

var msg string

func ValidateParam(param interface{}) string {
	_, validate := bindValidate()
	err := validate.Struct(param)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		for _, er := range errs {
			msg = er.Error()
		}
		return msg
	}
	return ""
}
func bindValidate() (locales.Translator, *validator.Validate) {
	validate := validator.New()
	e := en.New()

	return e, validate
}
