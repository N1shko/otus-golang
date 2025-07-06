package hw09structvalidator

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	BadStructRule struct {
		Version string `validate:"some:asd"`
	}
	BadStructUsage struct {
		Version string `validate:"len:asd"`
	}
	BadStructRegex struct {
		Version string `validate:"len:asd"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			App{"some1"},
			nil,
		},
		{
			User{
				"123",
				"Alex Johnson",
				17,
				"exampleNOATgmail.com",
				"stuffer",
				[]string{"88005553535", "12345678901"},
				json.RawMessage{},
			},
			ValidationErrors{
				ValidationError{
					"ID",
					fmt.Errorf("Value '123' not valid for rule 'len'"),
				},
				ValidationError{
					"Age",
					fmt.Errorf("Value '17' not valid for rule 'min'"),
				},
				ValidationError{
					"Role",
					fmt.Errorf("Value 'stuffer' not valid for rule 'in'"),
				},
				ValidationError{
					"Email",
					fmt.Errorf("Value 'exampleNOATgmail.com' not valid for rule 'regexp'"),
				},
			},
		},
	}

	wrongTests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			BadStructRule{"some1"},
			UnsupportedRule,
		},
		{
			BadStructUsage{"some1"},
			WrongRuleUsage,
		},
		{
			"somethingNotStrcut",
			ErrNotSupportedValue,
		},
		{
			123,
			ErrNotSupportedValue,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tests[i].in)
			assert.ElementsMatch(t, tests[i].expectedErr, err)
			_ = tt
		})
	}

	for i, tt := range wrongTests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(wrongTests[i].in)
			require.ErrorIs(t, err, wrongTests[i].expectedErr)
			_ = tt
		})
	}
}
