package genericsite

import (
	"testing"
)

func Test1(t *testing.T) {
	err := ConfirmationEmail("mindfish.me", "http://www.ikke.no/", "rodseth", "rodseth@gmail.com")
	if err != nil {
		t.Log("sending email failed")
	}
}
