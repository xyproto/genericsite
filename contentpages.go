package genericsite

import (
	"time"

	"github.com/drbawb/mustache"
	"github.com/hoisie/web"
	. "github.com/xyproto/onthefly"
	"github.com/xyproto/webhandle"
)

type ContentPage struct {
	GeneratedCSSurl          string
	ExtraCSSurls             []string
	JqueryJSurl              string
	Faviconurl               string
	BgImageURL               string
	StretchBackground        bool
	Title                    string
	Subtitle                 string
	ContentTitle             string
	ContentHTML              string
	HeaderJS                 string
	ContentJS                string
	SearchButtonText         string
	SearchURL                string
	FooterText               string
	BackgroundTextureURL     string
	DarkBackgroundTextureURL string
	FooterTextColor          string
	FooterColor              string
	UserState                *UserState
	RoundedLook              bool
	Url                      string
	ColorScheme              *ColorScheme
	SearchBox                bool
	GoogleFonts              []string
	CustomSansSerif          string
	CustomSerif              string
}

// Content page generator
type CPgen (func(userState *UserState) *ContentPage)

// A collection of ContentPages
type PageCollection []ContentPage

// Every input from the user must be intitially stored in a UserInput variable, not in a string!
// This is just to be aware of which data one should be careful with, and to keep it clean.
type UserInput string

type ColorScheme struct {
	Darkgray           string
	Nicecolor          string
	Menu_link          string
	Menu_hover         string
	Menu_active        string
	Default_background string
	TitleText          string
}

type BaseCP func(state *UserState) *ContentPage

type TemplateValueGeneratorFactory func(*UserState) webhandle.TemplateValueGenerator

// The default settings
// Do not publish this page directly, but use it as a basis for the other pages
func DefaultCP(userState *UserState) *ContentPage {
	var cp ContentPage
	cp.GeneratedCSSurl = "/css/style.css"
	cp.ExtraCSSurls = []string{"/css/menu.css"}
	// TODO: fallback to local jquery.min.js, google how
	cp.JqueryJSurl = "//ajax.googleapis.com/ajax/libs/jquery/2.0.0/jquery.min.js" // "/js/jquery-2.0.0.js"
	cp.Faviconurl = "/img/favicon.ico"
	cp.ContentTitle = "NOP"
	cp.ContentHTML = "NOP NOP NOP"
	cp.ContentJS = ""
	cp.HeaderJS = ""
	cp.SearchButtonText = "Search"
	cp.SearchURL = "/search"
	cp.SearchBox = true

	// http://wptheming.wpengine.netdna-cdn.com/wp-content/uploads/2010/04/gray-texture.jpg
	// TODO: Draw these two backgroundimages with a canvas instead
	cp.BackgroundTextureURL = "" // "/img/gray.jpg"
	// http://turbo.designwoop.com/uploads/2012/03/16_free_subtle_textures_subtle_dark_vertical.jpg
	cp.DarkBackgroundTextureURL = "/img/darkgray.jpg"

	cp.FooterColor = "black"
	cp.FooterTextColor = "#303040"

	cp.FooterText = "NOP"

	cp.UserState = userState
	cp.RoundedLook = false

	cp.Url = "/" // To be filled in when published

	// The default color scheme
	var cs ColorScheme
	cs.Darkgray = "#202020"
	cs.Nicecolor = "#5080D0"   // nice blue
	cs.Menu_link = "#c0c0c0"   // light gray
	cs.Menu_hover = "#efefe0"  // light gray, somewhat yellow
	cs.Menu_active = "#ffffff" // white
	cs.Default_background = "#000030"
	cs.TitleText = "#303030"

	cp.ColorScheme = &cs

	cp.GoogleFonts = []string{"Armata", "IM Fell English SC"}
	cp.CustomSansSerif = "" // Use the default sans serif
	cp.CustomSerif = "IM Fell English SC"

	return &cp
}

func genericPageBuilder(cp *ContentPage) *Page {
	// TODO: Record the time from one step out, because content may be generated and inserted into this generated conten
	startTime := time.Now()

	page := NewHTML5Page(cp.Title + " " + cp.Subtitle)

	page.LinkToCSS(cp.GeneratedCSSurl)
	for _, cssurl := range cp.ExtraCSSurls {
		page.LinkToCSS(cssurl)
	}
	page.LinkToJS(cp.JqueryJSurl)
	page.LinkToFavicon(cp.Faviconurl)

	AddHeader(page, cp.HeaderJS)
	AddGoogleFonts(page, cp.GoogleFonts)
	AddBodyStyle(page, cp.BgImageURL, cp.StretchBackground)
	AddTopBox(page, cp.Title, cp.Subtitle, cp.SearchURL, cp.SearchButtonText, cp.BackgroundTextureURL, cp.RoundedLook, cp.ColorScheme, cp.SearchBox)

	// TODO: Move the menubox into the TopBox

	AddMenuBox(page, cp.DarkBackgroundTextureURL, cp.CustomSansSerif)

	AddContent(page, cp.ContentTitle, cp.ContentHTML+DocumentReadyJS(cp.ContentJS))

	elapsed := time.Since(startTime)
	AddFooter(page, cp.FooterText, cp.FooterTextColor, cp.FooterColor, elapsed)

	return page
}

// Publish a list of ContentPaages, a colorscheme and template content
func PublishCPs(userState *UserState, pc PageCollection, cs *ColorScheme, tvgf TemplateValueGeneratorFactory, cssurl string) {
	// For each content page in the page collection
	for _, cp := range pc {
		// TODO: different css urls for all of these?
		cp.Pub(userState, cp.Url, cssurl, cs, tvgf(userState))
	}
}

// Some Engines like Admin must be served separately
// jquerypath is ie "/js/jquery.2.0.0.js", will then serve the file at static/js/jquery.2.0.0.js
func ServeSite(basecp BaseCP, userState *UserState, cps PageCollection, tvgf TemplateValueGeneratorFactory, jquerypath string) {
	cs := basecp(userState).ColorScheme
	PublishCPs(userState, cps, cs, tvgf, "/css/menu.css")

	// TODO: Add fallback to this local version
	webhandle.Publish(jquerypath, "static"+jquerypath, true)

	// TODO: Generate these
	webhandle.Publish("/robots.txt", "static/various/robots.txt", false)
	webhandle.Publish("/sitemap_index.xml", "static/various/sitemap_index.xml", false)
}

// CSS for the menu, and a bit more
func GenerateMenuCSS(state *UserState, stretchBackground bool, cs *ColorScheme) webhandle.SimpleContextHandle {
	return func(ctx *web.Context) string {
		ctx.ContentType("css")

		// one of the extra css files that are loaded after the main style
		retval := mustache.RenderFile("/usr/share/templates/menustyle.tmpl", cs)

		// The load order of background-color, background-size and background-image
		// is actually significant in some browsers! Do not reorder lightly.
		if stretchBackground {
			retval = "body {\nbackground-color: " + cs.Default_background + ";\nbackground-size: cover;\n}\n" + retval
		} else {
			retval = "body {\nbackground-color: " + cs.Default_background + ";\n}\n" + retval
		}
		retval += ".titletext { display: inline; }"
		return retval
	}
}

// Make an html and css page available
func (cp *ContentPage) Pub(userState *UserState, url, cssurl string, cs *ColorScheme, tvg webhandle.TemplateValueGenerator) {
	genericpage := genericPageBuilder(cp)
	web.Get(url, webhandle.GenerateHTMLwithTemplate(genericpage, tvg))
	web.Get(cp.GeneratedCSSurl, webhandle.GenerateCSS(genericpage))
	web.Get(cssurl, GenerateMenuCSS(userState, cp.StretchBackground, cs))
}

// TODO: Write a function for rendering a StandaloneTag inside a Page by the use of template {{{placeholders}}

// Render a page by inserting data at the {{{placeholders}}} for both html and css
func RenderPage(page *Page, templateContents map[string]string) (string, string) {
	// Note that the whitespace formatting of the generated html matter for the menu layout!
	return mustache.Render(page.String(), templateContents), mustache.Render(page.GetCSS(), templateContents)
}

// Wrap a lonely string in an entire webpage
func (cp *ContentPage) Surround(s string, templateContents map[string]string) (string, string) {
	cp.ContentHTML = s
	page := genericPageBuilder(cp)
	return RenderPage(page, templateContents)
}

// Uses a given WebHandle as the contents for the the ContentPage contents
func (cp *ContentPage) WrapWebHandle(wh webhandle.WebHandle, tvg webhandle.TemplateValueGenerator) webhandle.WebHandle {
	return func(ctx *web.Context, val string) string {
		html, css := cp.Surround(wh(ctx, val), tvg(ctx))
		web.Get(cp.GeneratedCSSurl, css)
		return html
	}
}

// Uses a given SimpleContextHandle as the contents for the the ContentPage contents
func (cp *ContentPage) WrapSimpleContextHandle(sch webhandle.SimpleContextHandle, tvg webhandle.TemplateValueGenerator) webhandle.SimpleContextHandle {
	return func(ctx *web.Context) string {
		html, css := cp.Surround(sch(ctx), tvg(ctx))
		web.Get(cp.GeneratedCSSurl, css)
		return html
	}
}
