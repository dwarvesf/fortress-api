package request

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisCustomValidators(engine *gin.Engine) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		regisCrossFieldGte(v)
	}
}

func regisCrossFieldGte(v *validator.Validate) {
	v.RegisterValidation("fieldgte", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		kind := field.Kind()
		params := strings.Split(fl.Param(), " ")

		currentField, currentKind, _, found := fl.GetStructFieldOKAdvanced2(fl.Parent(), params[0])

		if !found || currentKind != kind {
			return false
		}

		gteValue, err := strconv.ParseInt(params[1], 10, 64)
		if err != nil {
			return false
		}

		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int()-currentField.Int() >= gteValue
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return field.Uint()-currentField.Uint() >= uint64(gteValue)
		case reflect.Float32, reflect.Float64:
			return field.Float()-currentField.Float() >= float64(gteValue)
		}

		return len(field.String())-len(currentField.String()) >= int(gteValue)
	})
}
