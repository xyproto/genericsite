package genericsite

import (
	"strings"

	. "github.com/xyproto/browserspeak"
	"github.com/xyproto/web"
)

// Generate tags for the menu based on a list of "MenuDescription:/menu/url"
func MenuSnippet(links []string) *Page {
	var a, li, sep *Tag
	var text, url, firstword string

	page, ul := CowboyTag("ul")
	ul.AddStyle("list-style-type", "none")
	ul.AddStyle("float", "left")
	ul.AddStyle("margin", "0")

	for i, text_url := range links {

		text, url = ColonSplit(text_url)

		firstword = text
		if strings.Contains(text, " ") {
			// Get the first word of the menu text
			firstword = strings.SplitN(text, " ", 2)[0]
		}

		li = ul.AddNewTag("li")

		// TODO: Make sure not duplicate ids are added for two menu entries named "Hi there" and "Hi you"
		menuId := "menu" + firstword
		li.AddAttr("id", menuId)

		// All menu entries are now hidden by default!
		//li.AddStyle("display", "none")
		//li.AddStyle("display", "inline")
		li.SansSerif()
		//li.CustomSansSerif("Armata")

		// For every element, but not the first one
		if i > 0 {
			// Insert a '|' character in a div
			sep = li.AddNewTag("div")
			sep.AddContent("|")
			sep.AddAttr("class", "separator")
		}

		a = li.AddNewTag("a")
		a.AddAttr("class", "menulink")
		a.AddAttr("href", url)
		a.AddContent(text)

	}

	sep.AddStyle("display", "inline")
	sep.AddStyle("color", "#a0a0a0")

	return page

}

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
// Don't laugh, sometimes all you need is a FactoryGenerator
func DynamicMenuFactoryGenerator(currentMenuID string, usercontent []string) TemplateValueGeneratorFactory {
	return func(state *UserState) TemplateValueGenerator {
		return func(ctx *web.Context) TemplateValues {

			// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			//
			// The whole point is to hide those links that are not to be shown
			//
			// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

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
