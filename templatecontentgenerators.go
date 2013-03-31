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

//func DynamicMenuFactory(state *UserState) TemplateValueGenerator {
//	return func(ctx *web.Context) TemplateValues {
//		// Can generate the menu based on both the user state and the web context here
//		return TemplateValues{"menu": "<div style='margin-left: 3em;'><a href='/login'>Login</a> | <a href='/register'>Register</a></div>"}
//	}
//}


//func MenuEntry(url, text string) string {
func MenuEntry(id string) string {
	text := strings.Title(id)
	url := "/" + strings.ToLower(id)
	// TODO: Remove this special case once the menu generation has improved
	if url == "/overview" {
		url = "/"
	}
	return "<a href='" + url + "'>" + text + "</a> | "
}

func MenuStart() string {
	return "<div style='margin-left: 3em;'>"
}

func MenuEnd() string {
	return "</div>"
}

// TODO: Take the same parameters as the old menu generating code
// TODO: Put one if these in each engine then combine them somehow
func DynamicMenuFactoryGenerator(currentMenuID string, usercontent []string) TemplateValueGeneratorFactory {
	return func(state *UserState) TemplateValueGenerator {
		return func(ctx *web.Context) TemplateValues {

			retval := MenuStart()

			// Always show the Overview menu
			retval += MenuEntry("Overview")

			// If logged in, show Logout and the content
			if state.UserRights(ctx) {
				if currentMenuID != "Logout" {
					retval += MenuEntry("Logout")
				}

				// Also show the actual content
				for _, menuID := range usercontent {
					// Except the page we're on
					if menuID != currentMenuID {
						retval += MenuEntry(menuID)
					}
				}

				// Show admin content
				if state.AdminRights(ctx) {
					if currentMenuID != "Admin" {
						retval += MenuEntry("Admin")
					}
				}
			} else {
				// Only show Login and Register
				if currentMenuID != "Login" {
					retval += MenuEntry("Login")
				}
				if currentMenuID != "Register" {
					retval += MenuEntry("Register")
				}
			}
			retval += MenuEnd()

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
