package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var ErrNotSupportedValue = errors.New("Provided input is not struct")
var UnsupportedRule = errors.New("This rule is unsupported. Valid rules are: len, regexp, in, min, max")
var WrongRuleUsage = errors.New("Wrong usage for rule.")
var UncompilableRegex = errors.New("This regex cant be compiled")

var integerKinds = map[reflect.Kind]struct{}{
	reflect.Int:     {},
	reflect.Int8:    {},
	reflect.Int16:   {},
	reflect.Int32:   {},
	reflect.Int64:   {},
	reflect.Uint:    {},
	reflect.Uint8:   {},
	reflect.Uint16:  {},
	reflect.Uint32:  {},
	reflect.Uint64:  {},
	reflect.Uintptr: {},
}

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	msgErr := make([]string, 0, len(v))
	for _, msg := range v {
		msgErr = append(msgErr, msg.Err.Error())
	}
	return strings.Join(msgErr, "; ")
}

func Validate(v interface{}) error {
	type_ := reflect.TypeOf(v)
	value_ := reflect.ValueOf(v)
	var errors ValidationErrors
	if type_.Kind() != reflect.Struct {
		return ErrNotSupportedValue
	}
	for i := 0; i < type_.NumField(); i++ {
		tag, ok := type_.Field(i).Tag.Lookup("validate")
		if !ok {
			continue
		}
		if type_.Field(i).Type.Kind() == reflect.Slice {
			for j := 0; j < value_.Field(i).Len(); j++ {
				err := ValidateField(value_.Field(i).Index(j), type_.Field(i).Name, tag, &errors)
				if err != nil {
					return err
				}
			}
		} else {
			err := ValidateField(value_.Field(i), type_.Field(i).Name, tag, &errors)
			if err != nil {
				return err
			}
		}

	}
	if len(errors) != 0 {
		return errors
	}
	return nil
}

func ValidateField(v reflect.Value, field string, tag string, errors *ValidationErrors) error {
	requirements := strings.Split(tag, "|")
	for i := range requirements {
		preparedRule := strings.Split(requirements[i], ":")
		ruleType := preparedRule[0]
		switch ruleType {
		case "len":
			if v.Kind() != reflect.String {
				return fmt.Errorf("len tag should be used with string. %w", WrongRuleUsage)
			}
			if len(preparedRule) > 2 {
				return fmt.Errorf("len tag supports only 1 argument like len:32. %w", WrongRuleUsage)
			}
			l, err := strconv.Atoi(preparedRule[1])
			if err != nil {
				return fmt.Errorf("len tag argument should be integer. %w", WrongRuleUsage)
			}
			if v.Len() != l {
				*errors = append(*errors, ValidationError{field, fmt.Errorf("Value '%v' not valid for rule '%s'", v, ruleType)})
			}
		case "regexp":
			if v.Kind() != reflect.String {
				return fmt.Errorf("regexp tag should be used with string. %w", WrongRuleUsage)
			}
			regex := requirements[i][8:]
			re, err := regexp.Compile(regex)
			if err != nil {
				return fmt.Errorf("Found regex %s. %w", regex, UncompilableRegex)
			}
			if !re.MatchString(string(v.String())) {
				*errors = append(*errors, ValidationError{field, fmt.Errorf("Value '%v' not valid for rule '%s'", v, ruleType)})
			}
		case "in":
			var input string
			regex := strings.ReplaceAll(preparedRule[1], ",", "|")
			re, err := regexp.Compile(string('^') + regex + string('$'))
			if err != nil {
				return fmt.Errorf("Found 'in' rule %s. %w", regex, UncompilableRegex)
			}
			switch v.Kind() {
			case reflect.String:
				input = v.String()
			case reflect.Int:
				input = strconv.Itoa(int(v.Int()))
			default:
				return fmt.Errorf("'in' tag unsupported kind: %s", v.Kind())
			}
			if find := re.FindStringSubmatch(input); len(find) == 0 {
				*errors = append(*errors, ValidationError{field, fmt.Errorf("Value '%v' not valid for rule '%s'", v, ruleType)})
			}
		case "min", "max":
			limit, err := strconv.Atoi(preparedRule[1])
			if err != nil {
				return fmt.Errorf("%s tag requires integer: %w", ruleType, WrongRuleUsage)
			}
			if _, ok := integerKinds[v.Kind()]; !ok {
				return fmt.Errorf("%s tag only for integer: %w", ruleType, WrongRuleUsage)
			}
			val := int(v.Int())
			if (ruleType == "min" && val < limit) || (ruleType == "max" && val > limit) {
				*errors = append(*errors, ValidationError{field, fmt.Errorf("Value '%v' not valid for rule '%s'", v, ruleType)})
			}
		default:
			return fmt.Errorf("Found rule %s. %w", ruleType, UnsupportedRule)
		}
	}
	return nil
}
