package validateapp

import "github.com/go-playground/validator"

type StructToValidate interface {
	FuncToImplement()
}

func ValidateStruct(s StructToValidate) error {
	valid := validator.New()
	return valid.Struct(s)
}
