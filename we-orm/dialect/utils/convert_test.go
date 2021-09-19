package utils

import "testing"

func TestConvert(t *testing.T) {
	str := "HelloWorld"
	if ToUpper(str) != "HELLOWORLD" {
		t.Fatal("failed to ToUpper")
	}
	if ToLower(str) != "helloworld" {
		t.Fatal("failed to ToLower")
	}
	if CamelCaseToUnderscore(str) != "hello_world" {
		t.Fatal("failed to CamelCaseToUnderscore")
	}
	if CamelCaseToUnderscore("helloWorld") != "hello_world" {
		t.Fatal("failed to CamelCaseToUnderscore")
	}
	str = "hello_world"
	if UnderscoreToLowerCamelCase(str) != "helloWorld" {
		t.Fatal("failed to UnderscoreToLowerCamelCase")
	}
	if UnderscoreToUpperCamelCase(str) != "HelloWorld" {
		t.Fatal("failed to UnderscoreToUpperCamelCase")
	}

}
