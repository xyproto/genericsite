package genericsite

import (
	"strings"

	"github.com/xyproto/web"
)

func Menu(links []string) *Page {
	var a, li, sep *Tag
	var text, url, firstword string

	page, li = CowboyTag("ul")
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

		li = page.AddNewTag("li")

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
