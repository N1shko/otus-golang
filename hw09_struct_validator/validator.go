package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotSupportedValue = errors.New("provided input is not struct")
	ErrUnsupportedRule   = errors.New("this rule is unsupported, valid rules are: len, regexp, in, min, max")
	ErrWrongRuleUsage    = errors.New("wrong usage for rule")
	ErrUncompilableRegex = errors.New("this regex cant be compiled")
)

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
	dataType := reflect.TypeOf(v)
	dataValue := reflect.ValueOf(v)
	var errors ValidationErrors
	if dataType.Kind() != reflect.Struct {
		return ErrNotSupportedValue
	}
	for i := 0; i < dataType.NumField(); i++ {
		tag, ok := dataType.Field(i).Tag.Lookup("validate")
		if !ok {
			continue
		}
		if dataType.Field(i).Type.Kind() == reflect.Slice {
			for j := 0; j < dataValue.Field(i).Len(); j++ {
				err := ValidateField(dataValue.Field(i).Index(j), dataType.Field(i).Name, tag, &errors)
				if err != nil {
					return err
				}
			}
		} else {
			err := ValidateField(dataValue.Field(i), dataType.Field(i).Name, tag, &errors)
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
			err := validateLen(v, field, preparedRule, errors)
			if err != nil {
				return err
			}
		case "regexp":
			err := validateRegexp(v, field, requirements[i], errors)
			if err != nil {
				return err
			}
		case "in":
			err := validateIn(v, field, preparedRule, errors)
			if err != nil {
				return err
			}
		case "min", "max":
			err := validateMinMax(v, field, ruleType, preparedRule, errors)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("found rule %s. %w", ruleType, ErrUnsupportedRule)
		}
	}
	return nil
}

func validateLen(v reflect.Value, field string, rule []string, errors *ValidationErrors) error {
	if v.Kind() != reflect.String {
		return fmt.Errorf("len tag should be used with string. %w", ErrWrongRuleUsage)
	}
	if len(rule) > 2 {
		return fmt.Errorf("len tag supports only 1 argument like len:32. %w", ErrWrongRuleUsage)
	}
	l, err := strconv.Atoi(rule[1])
	if err != nil {
		return fmt.Errorf("len tag argument should be integer. %w", ErrWrongRuleUsage)
	}
	if v.Len() != l {
		*errors = append(*errors, ValidationError{field, fmt.Errorf("value '%v' not valid for rule 'len'", v)})
	}
	return nil
}

func validateRegexp(v reflect.Value, field string, rule string, errors *ValidationErrors) error {
	if v.Kind() != reflect.String {
		return fmt.Errorf("regexp tag should be used with string. %w", ErrWrongRuleUsage)
	}
	regex := rule[8:]
	re, err := regexp.Compile(regex)
	if err != nil {
		return fmt.Errorf("found regex %s. %w", regex, ErrUncompilableRegex)
	}
	if !re.MatchString(v.String()) {
		*errors = append(*errors, ValidationError{field, fmt.Errorf("value '%v' not valid for rule 'regexp'", v)})
	}
	return nil
}

func validateIn(v reflect.Value, field string, rule []string, errors *ValidationErrors) error {
	var input string
	if len(rule) < 2 {
		return fmt.Errorf("in tag requires at least one value. %w", ErrWrongRuleUsage)
	}
	regex := "^" + strings.ReplaceAll(rule[1], ",", "|") + "$"
	re, err := regexp.Compile(regex)
	if err != nil {
		return fmt.Errorf("found 'in' rule %s. %w", regex, ErrUncompilableRegex)
	}

	switch v.Kind() {
	case reflect.String:
		input = v.String()
	case reflect.Int:
		input = strconv.Itoa(int(v.Int()))
	default:
		return fmt.Errorf("'in' tag unsupported kind: %s", v.Kind())
	}
	if !re.MatchString(input) {
		*errors = append(*errors, ValidationError{field, fmt.Errorf("value '%v' not valid for rule 'in'", v)})
	}
	return nil
}

func validateMinMax(v reflect.Value, field string, ruleType string, rule []string, errors *ValidationErrors) error {
	if len(rule) < 2 {
		return fmt.Errorf("%s tag requires one argument. %w", ruleType, ErrWrongRuleUsage)
	}
	limit, err := strconv.Atoi(rule[1])
	if err != nil {
		return fmt.Errorf("%s tag requires integer: %w", ruleType, ErrWrongRuleUsage)
	}
	if _, ok := integerKinds[v.Kind()]; !ok {
		return fmt.Errorf("%s tag only for integer: %w", ruleType, ErrWrongRuleUsage)
	}
	val := int(v.Int())
	if (ruleType == "min" && val < limit) || (ruleType == "max" && val > limit) {
		*errors = append(*errors, ValidationError{field, fmt.Errorf("value '%v' not valid for rule '%s'", v, ruleType)})
	}
	return nil
}
