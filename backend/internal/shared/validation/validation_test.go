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
