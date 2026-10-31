package fedex

import "testing"

func TestBaseURL(t *testing.T) {
	if got := BaseURL(EnvProduction); got != "https://apis.fedex.com" {
		t.Fatal(got)
	}
	if got := BaseURL(EnvSandbox); got != "https://apis-sandbox.fedex.com" {
		t.Fatal(got)
	}
}

func TestNewDefaults(t *testing.T) {
	c := New(Config{ClientID: "id", ClientSecret: "sec"})
	if c.Environment() != EnvSandbox {
		t.Fatal(c.Environment())
	}
}
