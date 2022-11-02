package stepconf

import (
	"fmt"
	"reflect"

	"github.com/bitrise-io/go-utils/colorstring"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Print the name of the struct with Title case in blue color with followed by a newline,
// then print all fields formatted as `- field name: field value` separated by newline.
func Print(config interface{}) {
	fmt.Print(toString(config))
}

func valueString(v reflect.Value) string {
	if v.Kind() != reflect.Ptr {
		if v.Kind() == reflect.String && v.Len() == 0 {
			return "<unset>"
		}
		return fmt.Sprintf("%v", v.Interface())
	}

	if !v.IsNil() {
		return fmt.Sprintf("%v", v.Elem().Interface())
	}

	return ""
}

// returns the name of the struct with Title case in blue color followed by a newline,
// then print all fields formatted as `- field name: field value` separated by newline.
func toString(config interface{}) string {
	v := reflect.ValueOf(config)
	t := reflect.TypeOf(config)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	titleCaseName := cases.Title(language.English, cases.NoLower).String(t.Name())
	str := fmt.Sprint(colorstring.Bluef("%s:\n", titleCaseName))
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var key, _ = parseTag(field.Tag.Get("env"))
		if key == "" {
			key = field.Name
		}
		str += fmt.Sprintf("- %s: %s\n", key, valueString(v.Field(i)))
	}

	return str
}
