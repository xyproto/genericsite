package genericsite

// Dismisses various attempts at accessing this site. Honeypot?

import (
	"fmt"

	"github.com/hoisie/web"
	"github.com/xyproto/instapage"
)

func Hello() string {
	msg := "Hi"
	return instapage.Message("Hello", msg)
}

func ParamExample(ctx *web.Context) string {
	return fmt.Sprintf("%v\n", ctx.Params)
}

func ServeForFun() {
	// These appeared in the log
	bogus := []string{"/signup", "/wp-login.php", "/join.php", "/register.php", "/profile.php", "/user/register/", "/tools/quicklogin.one", "/sign_up.html", "/profile.php", "/ucp.php", "/account/register.php", "/join_form.php", "/tiki-register.php", "/YaBB.cgi/", "/YaBB.pl/", "/member/register", "/signup.php", "/blogs/load/recent", "/member/join.php"}
	for _, location := range bogus {
		web.Get(location, Hello)
	}

	bogusParam := []string{"/index.php", "/viewtopic.php"}
	for _, location := range bogusParam {
		web.Get(location, ParamExample)
	}
}
