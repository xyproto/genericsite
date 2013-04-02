package genericsite

import (
	"strings"

	. "github.com/xyproto/browserspeak"
	"github.com/xyproto/web"
)

type MenuEntry struct {
	id string
	text string
	url string
}

type MenuEntries []*MenuEntry

// Generate a menu ID based on the first word in the menu text
func (me *MenuEntry) AutoId() {
	firstword := me.text
	if strings.Contains(me.text, " ") {
		// Get the first word of the menu text
		firstword = strings.SplitN(me.text, " ", 2)[0]
	}
	me.id = firstword
}

// Takes something like "Admin:/admin" and returns a *MenuEntry
func NewMenuEntry(text_and_url string) *MenuEntry {
	var me MenuEntry
	me.text, me.url = ColonSplit(text_and_url)
	me.AutoId()
	return &me
}

func Links2menuEntries(links []string) MenuEntries {
	menuEntries := make(MenuEntries, len(links))
	for i, text_and_url := range links {
		menuEntries[i] = NewMenuEntry(text_and_url)
	}
	return menuEntries
}

// Generate tags for the menu based on a list of "MenuDescription:/menu/url"
func MenuSnippet(menuEntries MenuEntries) *Page {
	var a, li, sep *Tag

	page, ul := CowboyTag("ul")
	ul.AddAttr("class", "menuList")
	//ul.AddStyle("list-style-type", "none")
	//ul.AddStyle("float", "left")
	//ul.AddStyle("margin", "0")

	for i, menuEntry := range menuEntries {

		li = ul.AddNewTag("li")
		li.AddAttr("class", "menuEntry")

		// TODO: Make sure not duplicate ids are added for two menu entries named "Hi there" and "Hi you"
		menuId := "menu" + menuEntry.id
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
		a.AddAttr("href", menuEntry.url)
		a.AddContent(menuEntry.text)

	}

	return page
}

// Checks if a *MenuEntry exists in a []*MenuEntry (MenuEntries)
func HasEntry(checkEntry *MenuEntry, menuEntries MenuEntries) bool {
	for _, menuEntry := range menuEntries {
		if menuEntry.url == checkEntry.url {
			return true
		}
	}
	return false
}

func AddIfNotAdded(url string, currentMenuURL string, filteredMenuEntries *MenuEntries, menuEntry *MenuEntry) {
	if currentMenuURL != url {
		if menuEntry.url == url {
			if !HasEntry(menuEntry, *filteredMenuEntries) {
				*filteredMenuEntries = append(*filteredMenuEntries, menuEntry)
			}
		}
	}
}


/* 
 * Functions that generate functions that generate content that can be used in templates.
 * type TemplateValues map[string]string
 * type TemplateValueGenerator func(*web.Context) TemplateValues
 * type TemplateValueGeneratorFactory func(*UserState) TemplateValueGenerator
 */
// TODO: Take the same parameters as the old menu generating code
// TODO: Put one if these in each engine then combine them somehow
// TODO: Check for the menyEntry.url first, then check the rights, not the other way around
// TODO: Fix and refactor this one
// TODO: Check the user status _once_, and the admin status _once_, then generate the menu
func DynamicMenuFactoryGenerator(currentMenuURL string, menuEntries MenuEntries) TemplateValueGeneratorFactory {
	return func(state *UserState) TemplateValueGenerator {
		return func(ctx *web.Context) TemplateValues {

			var filteredMenuEntries MenuEntries
			var logoutEntry *MenuEntry = nil

			// Build up filteredMenuEntries based on what should be shown or not
			for _, menuEntry := range menuEntries {

				// Don't add duplicates
				if HasEntry(menuEntry, filteredMenuEntries) {
					continue
				}

				// Add this one last
				if menuEntry.url == "/logout" {
					if state.UserRights(ctx) {
						logoutEntry = menuEntry
					}
					continue
				}

				// Always show the Overview menu
				AddIfNotAdded("/", currentMenuURL, &filteredMenuEntries, menuEntry)
				//if menuEntry.url == "/" {
				//	if !HasEntry(menuEntry, filteredMenuEntries) {
				//		filteredMenuEntries = append(filteredMenuEntries, menuEntry)
				//	}
				//}

				// If logged in, show Logout and the content
				if state.UserRights(ctx) {

					// Add every link except the current page we're on
					if menuEntry.url != currentMenuURL {
						if !HasEntry(menuEntry, filteredMenuEntries) {
							if (menuEntry.url != "/login") && (menuEntry.url != "/register") {
								filteredMenuEntries = append(filteredMenuEntries, menuEntry)
							}
						}
					}

					// Show admin content
					if state.AdminRights(ctx) {
						AddIfNotAdded("/admin", currentMenuURL, &filteredMenuEntries, menuEntry)
					}
				} else {
					// Only show Login and Register
					AddIfNotAdded("/login", currentMenuURL, &filteredMenuEntries, menuEntry)
					AddIfNotAdded("/register", currentMenuURL, &filteredMenuEntries, menuEntry)
				}

			}

			if logoutEntry != nil {
				AddIfNotAdded("/logout", "", &filteredMenuEntries, logoutEntry)
			}

			page := MenuSnippet(filteredMenuEntries)
			retval := page.String()

			// TODO: Return the CSS as well somehow
			//css := page.CSS()

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
