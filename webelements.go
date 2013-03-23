package genericsite

// Various elements of a webpage

import (
	"strings"
	"time"

	. "github.com/xyproto/browserspeak"
)

type Engine struct {
	state *UserState
}

func AddTopBox(page *Page, title, subtitle, searchURL, searchButtonText, backgroundTextureURL string, roundedLook bool, cs *ColorScheme) (*Tag, error) {
	body, err := page.GetTag("body")
	if err != nil {
		return nil, err
	}

	div := body.AddNewTag("div")
	div.AddAttr("id", "topbox")
	div.AddStyle("display", "block")
	div.AddStyle("width", "100%")
	div.AddStyle("margin", "0")
	div.AddStyle("padding", "0 0 1em 0")
	div.AddStyle("top", "0")
	div.AddStyle("left", "0")
	div.AddStyle("background-color", cs.Darkgray)
	div.AddStyle("position", "fixed")
	div.AddStyle("display", "block")

	titlebox := AddTitleBox(div, title, subtitle, cs)
	titlebox.AddAttr("id", "titlebox")
	titlebox.AddStyle("margin", "0 0 0 0")
	// Padding-top + height should be 5em, padding decides the position
	titlebox.AddStyle("padding", "1.8em 0 0 2.8em")
	titlebox.AddStyle("height", "3.2em")
	titlebox.AddStyle("width", "100%")
	titlebox.AddStyle("position", "fixed")
	titlebox.AddStyle("background-color", cs.Darkgray) // gray, could be a gradient
	titlebox.AddStyle("background", "url('"+backgroundTextureURL+"')")
	//titlebox.AddStyle("z-index", "2") // 2 is above the search box which is 1

	searchbox := AddSearchBox(titlebox, searchURL, searchButtonText, roundedLook)
	searchbox.AddAttr("id", "searchbox")
	searchbox.AddStyle("position", "relative")
	searchbox.AddStyle("float", "right")
	// The padding decides the position for this one
	searchbox.AddStyle("padding", "0 5em 0 0")
	searchbox.AddStyle("margin", "0")
	//searchbox.AddStyle("min-width", "10em")
	//searchbox.AddStyle("line-height", "10em")
	//searchbox.AddStyle("z-index", "1") // below the title

	return div, nil
}

// TODO: Place at the bottom of the content instead of at the bottom of the window
func AddFooter(page *Page, footerText, footerTextColor, footerColor string, elapsed time.Duration) (*Tag, error) {
	body, err := page.GetTag("body")
	if err != nil {
		return nil, err
	}
	div := body.AddNewTag("div")
	div.AddAttr("id", "notice")
	div.AddStyle("position", "fixed")
	div.AddStyle("bottom", "0")
	div.AddStyle("left", "0")
	div.AddStyle("width", "100%")
	div.AddStyle("display", "block")
	div.AddStyle("padding", "0")
	div.AddStyle("margin", "0")
	div.AddStyle("background-color", footerColor)
	div.AddStyle("font-size", "0.6em")
	div.AddStyle("text-align", "right")
	div.AddStyle("box-shadow", "1px -2px 3px rgba(0,0,0, .5)")

	innerdiv := div.AddNewTag("div")
	innerdiv.AddAttr("id", "innernotice")
	innerdiv.AddStyle("padding", "0 2em 0 0")
	innerdiv.AddStyle("margin", "0")
	innerdiv.AddStyle("color", footerTextColor)
	innerdiv.AddContent("Generated in " + elapsed.String() + " | " + footerText)

	return div, nil
}

func AddContent(page *Page, contentTitle, contentHTML string) (*Tag, error) {
	body, err := page.GetTag("body")
	if err != nil {
		return nil, err
	}

	div := body.AddNewTag("div")
	div.AddAttr("id", "content")
	div.AddStyle("z-index", "-1")
	div.AddStyle("color", "black") // content headline color
	div.AddStyle("min-height", "80%")
	div.AddStyle("min-width", "60%")
	div.AddStyle("float", "left")
	div.AddStyle("position", "relative")
	div.AddStyle("margin-left", "4%")
	div.AddStyle("margin-top", "9.5em")
	div.AddStyle("margin-right", "5em")
	div.AddStyle("padding-left", "4em")
	div.AddStyle("padding-right", "5em")
	div.AddStyle("padding-top", "1em")
	div.AddStyle("padding-bottom", "2em")
	div.AddStyle("background-color", "rgba(255,255,255,0.92)")                                                                               // light gray. Transparency with rgba() doesn't work in IE
	div.AddStyle("filter", "progid:DXImageTransform.Microsoft.gradient(GradientType=0,startColorstr='#dcffffff', endColorstr='#dcffffff');") // for transparency in IE

	div.AddStyle("text-align", "justify")
	div.RoundedBox()

	h2 := div.AddNewTag("h2")
	h2.AddAttr("id", "textheader")
	h2.AddContent(contentTitle)
	h2.CustomSansSerif("Armata")

	p := div.AddNewTag("p")
	p.AddAttr("id", "textparagraph")
	p.AddStyle("margin-top", "0.5em")
	//p.CustomSansSerif("Junge")
	p.SansSerif()
	p.AddStyle("font-size", "1.0em")
	p.AddStyle("color", "black") // content text color
	p.AddContent(contentHTML)

	return div, nil
}

// Add a search box to the page, actionURL is the url to use as a get action,
// buttonText is the text on the search button
func AddSearchBox(tag *Tag, actionURL, buttonText string, roundedLook bool) *Tag {

	div := tag.AddNewTag("div")
	div.AddAttr("id", "searchboxdiv")
	div.AddStyle("text-align", "right")
	div.AddStyle("display", "inline-block")

	form := div.AddNewTag("form")
	form.AddAttr("id", "search")
	form.AddAttr("method", "get")
	form.AddAttr("action", actionURL)

	innerDiv := form.AddNewTag("div")
	innerDiv.AddAttr("id", "innerdiv")
	innerDiv.AddStyle("overflow", "hidden")
	innerDiv.AddStyle("padding-right", "0.5em")
	innerDiv.AddStyle("display", "inline-block")

	inputText := innerDiv.AddNewTag("input")
	inputText.AddAttr("id", "inputtext")
	inputText.AddAttr("name", "q")
	inputText.AddAttr("size", "25")
	inputText.AddStyle("padding", "0.25em")
	inputText.CustomSansSerif("Armata")
	inputText.AddStyle("background-color", "#f0f0f0")
	if roundedLook {
		inputText.RoundedBox()
	} else {
		inputText.AddStyle("border", "none")
	}

	// inputButton := form.AddNewTag("input")
	// inputButton.AddAttr("id", "inputbutton")
	// // The position is in the margin
	// inputButton.AddStyle("margin", "0.08em 0 0 0.4em")
	// inputButton.AddStyle("padding", "0.2em 0.6em 0.2em 0.6em")
	// inputButton.AddStyle("float", "right")
	// inputButton.AddAttr("type", "submit")
	// inputButton.AddAttr("value", buttonText)
	// inputButton.SansSerif()
	// //inputButton.AddStyle("overflow", "hidden")
	// if roundedLook {
	// 	inputButton.RoundedBox()
	// } else {
	// 	inputButton.AddStyle("border", "none")
	// }

	return div
}

func AddTitleBox(tag *Tag, title, subtitle string, cs *ColorScheme) *Tag {

	div := tag.AddNewTag("div")
	div.AddAttr("id", "titlebox")
	div.AddStyle("display", "block")
	div.AddStyle("position", "fixed")

	word1 := title
	word2 := ""
	if strings.Contains(title, " ") {
		word1 = strings.SplitN(title, " ", 2)[0]
		word2 = strings.SplitN(title, " ", 2)[1]
	}

	a := div.AddNewTag("a")
	a.AddAttr("id", "homelink")
	a.AddAttr("href", "/")
	a.AddStyle("text-decoration", "none")

	font0 := a.AddNewTag("div")
	font0.AddAttr("id", "whitetitle")
	font0.AddStyle("color", "white")
	//font0.CustomSansSerif("Armata")
	font0.SansSerif()
	font0.AddStyle("font-size", "2.0em")
	font0.AddStyle("font-weight", "bolder")
	font0.AddContent(word1)

	font1 := a.AddNewTag("div")
	font1.AddAttr("id", "bluetitle")
	font1.AddStyle("color", cs.Nicecolor)
	//font1.CustomSansSerif("Armata")
	font1.SansSerif()
	font1.AddStyle("font-size", "2.0em")
	font1.AddStyle("font-weight", "bold")
	font1.AddStyle("overflow", "hidden")
	font1.AddContent(word2)

	font2 := a.AddNewTag("div")
	font2.AddAttr("id", "graytitle")
	font2.AddStyle("font-size", "0.5em")
	font2.AddStyle("color", "#707070")
	font2.CustomSansSerif("Armata")
	font2.AddStyle("font-size", "1.25em")
	font2.AddStyle("font-weight", "normal")
	font2.AddStyle("overflow", "hidden")
	font2.AddContent(subtitle)

	return div
}

// Takes a page and a colon-separated string slice of text:url, hiddenlinks is just a list of the url part
func AddMenuBox(page *Page, links, hiddenlinks []string, darkBackgroundTexture string) (*Tag, error) {
	body, err := page.GetTag("body")
	if err != nil {
		return nil, err
	}

	div := body.AddNewTag("div")
	div.AddAttr("id", "menubox")
	div.AddStyle("display", "block")
	div.AddStyle("width", "100%")
	div.AddStyle("margin", "0")
	div.AddStyle("padding", "0.1em 0 0.2em 0")
	div.AddStyle("position", "absolute")
	div.AddStyle("top", "5em")
	div.AddStyle("left", "0")
	div.AddStyle("background-color", "#0c0c0c") // dark gray, fallback
	div.AddStyle("background", "url('"+darkBackgroundTexture+"')")
	div.AddStyle("position", "fixed")
	div.AddStyle("box-shadow", "1px 3px 5px rgba(0,0,0, .8)")

	ul := div.AddNewTag("ul")
	ul.AddStyle("list-style-type", "none")
	ul.AddStyle("float", "left")
	ul.AddStyle("margin", "0")

	var a, li, sep *Tag
	var text, url, firstword string

	for i, text_url := range links {
		text, url = ColonSplit(text_url)

		firstword = text
		if strings.Contains(text, " ") {
			// Get the first word of the menu text
			firstword = strings.SplitN(text, " ", 2)[0]
		}

		li = ul.AddNewTag("li")

		// TODO: Make sure not duplicate ids are added for two menu entries named "Hi there" and "Hi you"
		li.AddAttr("id", "menu"+firstword)
		li.AddStyle("display", "inline")
		li.SansSerif()
		//li.CustomSansSerif("Armata")

		// Hide the menu items with matching urls
		for _, val := range hiddenlinks {
			if val == url {
				li.AddStyle("display", "none")
				break
			}
		}

		// For every element, but not the first one
		if i > 0 {
			// Insert a '|' character in a div
			sep = li.AddNewTag("div")
			sep.AddContent("|")
		}

		a = li.AddNewTag("a")
		a.AddAttr("class", "menulink")
		a.AddAttr("href", url)
		a.AddContent(text)
	}

	// For Login, Logout and Register
	// TODO: Implement this method too
	ul.AddLastContent("{{{yihaa}}}")

	sep.AddStyle("display", "inline")
	sep.AddStyle("color", "#a0a0a0")

	return div, nil
}
