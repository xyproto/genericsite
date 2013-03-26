package genericsite

import (
	"github.com/xyproto/web"
)

func show(id string) string {
	return "#menu" + id + " { display: inline; }\n"
}

func hide(id string) string {
	return "#menu" + id + " { display: none; }\n"
}

// Assume everything is hidden by default, menuID is for example "Login"
func MenuCSS(currentMenuID string, state *UserState, ctx *web.Context, usercontent []string) string {
	var css string
	// If logged in, show Logout and the content
	if state.UserRights(ctx) {
		css += show("Logout")

		// Also show the actual content
		for _, menuID := range usercontent {
			// Except the page we're on
			if menuID != currentMenuID {
				css += show(menuID)
			}
		}

		// Show admin content
		if state.AdminRights(ctx) {
			if currentMenuID != "Admin" {
				css += show("Admin")
			}
		}
	} else {
		// Only show Login and Register
		css += show("Login")
		css += show("Register")
	}
	// Always show the Overview menu
	css += show("Overview")
	return css
}
