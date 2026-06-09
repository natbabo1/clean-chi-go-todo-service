package validator

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	return &Validator{v: validator.New()}
}

// Validate returns a map of field→message, or nil if valid.
func (vl *Validator) Validate(s any) map[string]string {
	err := vl.v.Struct(s)
	if err == nil {
		return nil
	}
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{"_": err.Error()}
	}
	out := make(map[string]string, len(errs))
	for _, e := range errs {
		out[e.Field()] = e.Tag()
	}
	return out
}
