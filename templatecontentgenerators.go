package genericsite

import (
	"strings"

	. "github.com/xyproto/browserspeak"
	"github.com/xyproto/web"
)

/* 
 * Functions that generate functions that generate content that can be used in templates.
 *
 * From browserspeak, a reminder:
 *
 * type TemplateValues map[string]string
 * type TemplateValueGenerator func(*web.Context) TemplateValues
 * type TemplateValueGeneratorFactory func(*UserState) TemplateValueGenerator
 */

func DynamicMenuFactory(state *UserState) TemplateValueGenerator {
	return func(ctx *web.Context) TemplateValues {
		// Can generate the menu based on both the user state and the web context here
		return TemplateValues{"menu": "<div style='margin-left: 3em;'><a href='/login'>Login</a> | <a href='/register'>Register</a></div>"}
	}
}


//func AddMenuEntry(url, text string) string {
func AddMenuEntry(id string) string {
	text := strings.Title(id)
	url := "/" + id
	return "<div><a href='" + url + "'>" + text + "</a> | </div>"
}

// TODO: Put one if these in each engine then combine them somehow
func DynamicMenuFactoryGenerator(state *UserState, currentMenuID string, usercontent []string) TemplateValueGeneratorFactory {
	return func(state *UserState) TemplateValueGenerator {
		return func(ctx *web.Context) TemplateValues {

			var retval string

			// If logged in, show Logout and the content
			if state.UserRights(ctx) {
				if currentMenuID != "Logout" {
					retval += AddMenuEntry("Logout")
				}

				// Also show the actual content
				for _, menuID := range usercontent {
					// Except the page we're on
					if menuID != currentMenuID {
						retval += AddMenuEntry(menuID)
					}
				}

				// Show admin content
				if state.AdminRights(ctx) {
					if currentMenuID != "Admin" {
						retval += AddMenuEntry("Admin")
					}
				}
			} else {
				// Only show Login and Register
				if currentMenuID != "Login" {
					retval += AddMenuEntry("Login")
				}
				if currentMenuID != "Register" {
					retval += AddMenuEntry("Register")
				}
			}
			// Always show the Overview menu
			retval += AddMenuEntry("Overview")

			return TemplateValues{"menu": retval}
		}
	}
}

// Combines two TemplateValueGenerators into one TemplateValueGenerator by adding the strings per key
func TemplateValueGeneratorCombinator(tvg1, tvg2 TemplateValueGenerator) TemplateValueGenerator {
	return func(ctx *web.Context) TemplateValues {
		tv1 := tvg1(ctx)
		tv2 := tvg2(ctx)
		for key, value := range tv2 {
			// TODO: Check if key exists in tv1 first
			tv1[key] += value
		}
		return tv1
	}
}
