// Package validator provides the small, explicit validation surface used by HTTP handlers.
package validator

import (
	"fmt"
	"reflect"
	"strings"
)

// Validate supports the DTO tags used by this API: required, email, min, max and oneof.
func Validate(v any) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		tag := rt.Field(i).Tag.Get("validate")
		if tag == "" {
			continue
		}
		value := rv.Field(i)
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				value = reflect.Value{}
			} else {
				value = value.Elem()
			}
		}
		text := ""
		if value.IsValid() && value.Kind() == reflect.String {
			text = value.String()
		}
		for _, rule := range strings.Split(tag, ",") {
			switch {
			case rule == "required" && strings.TrimSpace(text) == "":
				return fmt.Errorf("%s is required", rt.Field(i).Name)
			case rule == "email" && (!strings.Contains(text, "@") || strings.HasPrefix(text, "@") || strings.HasSuffix(text, "@")):
				return fmt.Errorf("%s must be a valid email", rt.Field(i).Name)
			case strings.HasPrefix(rule, "min="):
				var n int
				fmt.Sscanf(strings.TrimPrefix(rule, "min="), "%d", &n)
				if len(text) < n {
					return fmt.Errorf("%s is too short", rt.Field(i).Name)
				}
			case strings.HasPrefix(rule, "max="):
				var n int
				fmt.Sscanf(strings.TrimPrefix(rule, "max="), "%d", &n)
				if len(text) > n {
					return fmt.Errorf("%s is too long", rt.Field(i).Name)
				}
			case strings.HasPrefix(rule, "oneof="):
				allowed := strings.Fields(strings.TrimPrefix(rule, "oneof="))
				found := false
				for _, a := range allowed {
					found = found || text == a
				}
				if !found {
					return fmt.Errorf("%s has an invalid value", rt.Field(i).Name)
				}
			}
		}
	}
	return nil
}
