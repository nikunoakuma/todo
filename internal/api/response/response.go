package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
	"todo/internal/models"
)

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() models.Response {
	return models.Response{
		Status: StatusOK,
	}
}

func Err(msg string) models.Response {
	return models.Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationErrorsResponse(errs validator.ValidationErrors) models.Response {
	var msgErrs = make([]string, 0, len(errs))

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			msgErrs = append(msgErrs, fmt.Sprintf("%s is a required field", err.Field()))
		default:
			msgErrs = append(msgErrs, fmt.Sprintf("field %s isn't valid"), err.Field())
		}
	}

	return Err(strings.Join(msgErrs, ", "))
}
