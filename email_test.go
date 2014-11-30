package genericsite

import (
	"testing"
)

func Test1(t *testing.T) {
	err := ConfirmationEmail("origin_domain.com", "http://somewhe.re/", "username", "username@somewhe.re")
	if err != nil {
		t.Log("sending email failed")
	}
}
