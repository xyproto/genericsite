package genericsite

import (
	"testing"
)

func Test1(t *testing.T) {
	err := ConfirmationEmail("archlinux.no", "http://www.ikke.no/", "rodseth", "jeje@archlinux.no")
	if err != nil {
		t.Log("sending email failed")
	}
}
