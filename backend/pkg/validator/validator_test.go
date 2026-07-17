package validator

import "testing"

type validationFixture struct {
	Name     string  `validate:"required,min=2,max=10"`
	Email    string  `validate:"required,email"`
	Role     string  `validate:"required,oneof=user admin"`
	Image    *string `validate:"omitempty,url,max=100"`
	Password string  `validate:"required,min=8"`
	New      string  `validate:"required,min=8,nefield=Password"`
}

func validFixture() validationFixture {
	return validationFixture{Name: "User", Email: "user@example.com", Role: "user", Password: "password-1", New: "password-2"}
}

func TestValidateAcceptsValidValues(t *testing.T) {
	fixture := validFixture()
	image := "https://example.com/avatar.png"
	fixture.Image = &image
	if err := Validate(fixture); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateAllowsOmittedOptionalValue(t *testing.T) {
	if err := Validate(validFixture()); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsInvalidRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*validationFixture)
	}{
		{name: "required", mutate: func(v *validationFixture) { v.Name = "" }},
		{name: "email", mutate: func(v *validationFixture) { v.Email = "invalid" }},
		{name: "oneof", mutate: func(v *validationFixture) { v.Role = "owner" }},
		{name: "minimum", mutate: func(v *validationFixture) { v.Name = "A" }},
		{name: "maximum", mutate: func(v *validationFixture) { v.Name = "A very long name" }},
		{name: "url", mutate: func(v *validationFixture) { image := "not-a-url"; v.Image = &image }},
		{name: "different field", mutate: func(v *validationFixture) { v.New = v.Password }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := validFixture()
			tt.mutate(&fixture)
			if err := Validate(fixture); err == nil {
				t.Fatal("Validate() should reject the value")
			}
		})
	}
}
