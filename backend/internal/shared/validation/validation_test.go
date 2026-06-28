package validation_test

import (
	"testing"

	"github.com/stayflow/stayflow-track/internal/shared/validation"
)

type TestStruct struct {
	Name  string `validate:"required,min=2,max=50"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=1,max=150"`
}

func TestValidate_Success(t *testing.T) {
	input := TestStruct{
		Name:  "John",
		Email: "john@example.com",
		Age:   25,
	}

	errs := validation.Validate(input)
	if errs != nil {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestValidate_Required(t *testing.T) {
	input := TestStruct{
		Name:  "",
		Email: "",
		Age:   25,
	}

	errs := validation.Validate(input)
	if errs == nil {
		t.Fatal("expected validation errors")
	}

	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(errs))
	}
}

func TestValidate_Email(t *testing.T) {
	input := TestStruct{
		Name:  "John",
		Email: "not-an-email",
		Age:   25,
	}

	errs := validation.Validate(input)
	if errs == nil {
		t.Fatal("expected validation errors for invalid email")
	}

	found := false
	for _, e := range errs {
		if e.Field == "email" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error for email field")
	}
}

func TestValidate_MinLength(t *testing.T) {
	input := TestStruct{
		Name:  "J",
		Email: "john@example.com",
		Age:   25,
	}

	errs := validation.Validate(input)
	if errs == nil {
		t.Fatal("expected validation error for min length")
	}
}

type MaxLengthStruct struct {
	Name string `validate:"required,max=5"`
}

func TestValidate_MaxLength(t *testing.T) {
	input := MaxLengthStruct{
		Name: "TooLongName",
	}

	errs := validation.Validate(input)
	if errs == nil {
		t.Fatal("expected validation error for max length")
	}

	found := false
	for _, e := range errs {
		if e.Field == "name" && e.Message == "must be at most 5 characters" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected max length error message, got: %v", errs)
	}
}

type OneOfStruct struct {
	Status string `validate:"required,oneof=active inactive"`
}

func TestValidate_OneOf(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		input := OneOfStruct{Status: "active"}
		errs := validation.Validate(input)
		if errs != nil {
			t.Fatalf("expected no errors, got: %v", errs)
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		input := OneOfStruct{Status: "deleted"}
		errs := validation.Validate(input)
		if errs == nil {
			t.Fatal("expected validation error for oneof")
		}
		found := false
		for _, e := range errs {
			if e.Field == "status" && e.Message == "must be one of: active inactive" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected oneof error message, got: %v", errs)
		}
	})
}

type UUIDStruct struct {
	ID string `validate:"required,uuid"`
}

func TestValidate_UUID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		input := UUIDStruct{ID: "123e4567-e89b-12d3-a456-426614174000"}
		errs := validation.Validate(input)
		if errs != nil {
			t.Fatalf("expected no errors for valid UUID, got: %v", errs)
		}
	})

	t.Run("invalid UUID", func(t *testing.T) {
		input := UUIDStruct{ID: "not-a-uuid"}
		errs := validation.Validate(input)
		if errs == nil {
			t.Fatal("expected validation error for invalid UUID")
		}
	})
}

type URLStruct struct {
	Website string `validate:"required,url"`
}

func TestValidate_URL(t *testing.T) {
	t.Run("valid URL", func(t *testing.T) {
		input := URLStruct{Website: "https://example.com"}
		errs := validation.Validate(input)
		if errs != nil {
			t.Fatalf("expected no errors, got: %v", errs)
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		input := URLStruct{Website: "not-a-url"}
		errs := validation.Validate(input)
		if errs == nil {
			t.Fatal("expected validation error for invalid URL")
		}
	})
}

type CamelCaseStruct struct {
	FirstName string `validate:"required,min=2"`
	LastName  string `validate:"required,min=2"`
	PhoneNum  string `validate:"required"`
}

func TestValidate_SnakeCaseFieldNames(t *testing.T) {
	input := CamelCaseStruct{
		FirstName: "",
		LastName:  "",
		PhoneNum:  "",
	}

	errs := validation.Validate(input)
	if errs == nil {
		t.Fatal("expected validation errors")
	}

	fieldNames := make(map[string]bool)
	for _, e := range errs {
		fieldNames[e.Field] = true
	}

	if !fieldNames["first_name"] {
		t.Error("expected field name 'first_name' (snake_case)")
	}
	if !fieldNames["last_name"] {
		t.Error("expected field name 'last_name' (snake_case)")
	}
	if !fieldNames["phone_num"] {
		t.Error("expected field name 'phone_num' (snake_case)")
	}
}

func TestValidate_ErrorMessages(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		field   string
		message string
	}{
		{
			"required message",
			struct {
				Name string `validate:"required"`
			}{},
			"name",
			"this field is required",
		},
		{
			"email message",
			struct {
				Email string `validate:"required,email"`
			}{Email: "bad"},
			"email",
			"must be a valid email address",
		},
		{
			"min message",
			struct {
				Pass string `validate:"required,min=8"`
			}{Pass: "abc"},
			"pass",
			"must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validation.Validate(tt.input)
			if errs == nil {
				t.Fatal("expected errors")
			}
			found := false
			for _, e := range errs {
				if e.Field == tt.field && e.Message == tt.message {
					found = true
				}
			}
			if !found {
				t.Errorf("expected field=%q message=%q, got: %v", tt.field, tt.message, errs)
			}
		})
	}
}
