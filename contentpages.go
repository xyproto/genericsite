package genericsite

import (
	"mime"
	"time"

	"github.com/drbawb/mustache"
	. "github.com/xyproto/browserspeak"
	"github.com/xyproto/web"
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
	Links                    []string
	HiddenMenuIDs            []string
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
}

// Content page generator
type CPgen (func(userState *UserState) *ContentPage)

// A collection of ContentPages
type PageCollection []ContentPage

// Every input from the user must be intitially stored in a UserInput variable, not in a string!
// This is just to be aware of which data one should be careful with, and to keep it clean.
type UserInput string

// Engine is a sub-system, a collection of sub-pages, that can be served
type Engine interface {
	GetState() *UserState
	SetState(*UserState)
	GenerateCSS(*ColorScheme) SimpleContextHandle
	ShowMenu(url string, ctx *web.Context) bool // Show menu for this engine?
	ServePages(BaseCP, *ColorScheme, map[string]string)
	ServeSystem()
}

type ColorScheme struct {
	Darkgray           string
	Nicecolor          string
	Menu_link          string
	Menu_hover         string
	Menu_active        string
	Default_background string
}

const (
	JQUERY_VERSION = "1.9.1"
)

// The default settings
// Do not publish this page directly, but use it as a basis for the other pages
func DefaultCP(userState *UserState) *ContentPage {
	var cp ContentPage
	cp.GeneratedCSSurl = "/css/style.css"
	cp.ExtraCSSurls = []string{"/css/extra.css"}
	// TODO: fallback to local jquery.min.js, google how
	cp.JqueryJSurl = "//ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js" // "/js/jquery-1.9.1.js"
	cp.Faviconurl = "/favicon.ico"
	cp.Links = []string{"Overview:/", "Login:/login", "Logout:/logout", "Register:/register", "Admin:/admin"}
	cp.ContentTitle = "NOP"
	cp.ContentHTML = "NOP NOP NOP"
	cp.ContentJS = ""
	cp.HeaderJS = ""
	cp.SearchButtonText = "Search"
	cp.SearchURL = "/search"

	// http://wptheming.wpengine.netdna-cdn.com/wp-content/uploads/2010/04/gray-texture.jpg
	// TODO: Draw these two backgroundimages with a canvas instead
	cp.BackgroundTextureURL = "/img/gray.jpg"
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
	cp.ColorScheme = &cs

	// Menus that are hidden (not generated) by default
	cp.HiddenMenuIDs = []string{"menuLogin", "menuLogout", "menuRegister"}

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
	AddGoogleFonts(page, []string{"Armata"}) //, "Junge"})
	AddBodyStyle(page, cp.BgImageURL, cp.StretchBackground)
	AddTopBox(page, cp.Title, cp.Subtitle, cp.SearchURL, cp.SearchButtonText, cp.BackgroundTextureURL, cp.RoundedLook, cp.ColorScheme)

	// TODO:
	// Use something dynamic to add or remove /login and /register depending on the login status
	// The login status can be fetched over AJAX or REST or something.

	// TODO: Move the menubox into the TopBox

	// TODO: Do this with templates or CSS instead
	// Hide login/logout/register by default
	AddMenuBox(page, cp.Links, cp.HiddenMenuIDs, cp.DarkBackgroundTextureURL)

	AddContent(page, cp.ContentTitle, cp.ContentHTML+BodyJS(cp.ContentJS))

	elapsed := time.Since(startTime)
	AddFooter(page, cp.FooterText, cp.FooterTextColor, cp.FooterColor, elapsed)

	return page
}

// Publish a list of ContentPaages, a colorscheme and template content
func PublishCPs(pc PageCollection, cs *ColorScheme, tp map[string]string, cssurl string) {
	// For each content page in the page collection
	for _, cp := range pc {
		// TODO: different css urls for all of these?
		cp.Pub(cp.Url, cssurl, cs, tp)
	}
}

type BaseCP func(state *UserState) *ContentPage

// Some Engines like Admin must be served separately
func ServeSite(basecp BaseCP, userState *UserState, cps PageCollection, tp map[string]string) {
	// Add pages for login, logout and register
	cps = append(cps, *LoginCP(basecp, userState, "/login"))
	cps = append(cps, *RegisterCP(basecp, userState, "/register"))

	cs := basecp(userState).ColorScheme
	PublishCPs(cps, cs, tp, "/css/extra.css")

	ServeSearchPages(basecp, userState, cps, cs, tp)

	//ServeAdminPages(basecp, userState, cs, tp)

	// TODO: Add fallback to this local version
	Publish("/js/jquery-"+JQUERY_VERSION+".js", "static/js/jquery-"+JQUERY_VERSION+".js", true)

	// TODO: Generate these
	Publish("/robots.txt", "static/various/robots.txt", false)
	Publish("/sitemap_index.xml", "static/various/sitemap_index.xml", false)
	Publish("/favicon.ico", "static/img/favicon.ico", false)
}

func GenerateMenuCSS(stretchBackground bool, cs *ColorScheme) SimpleContextHandle {
	return func(ctx *web.Context) string {
		ctx.ContentType("css")
		// one of the extra css files that are loaded after the main style
		retval := `
a {
  text-decoration: none;
  color: #303030;
  font-weight: regular;
}
a:link {color:` + cs.Menu_link + `;}
a:visited {color:` + cs.Menu_link + `;}
a:hover {color:` + cs.Menu_hover + `;}
a:active {color:` + cs.Menu_active + `;}
`
		// The load order of background-color, background-size and background-image
		// is actually significant in Chrome! Do not reorder lightly!
		if stretchBackground {
			retval = "body {\nbackground-color: " + cs.Default_background + ";\nbackground-size: cover;\n}\n" + retval
		} else {
			retval = "body {\nbackground-color: " + cs.Default_background + ";\n}\n" + retval
		}
		return retval
	}
}

// Make an html and css page available
func (cp *ContentPage) Pub(url, cssurl string, cs *ColorScheme, templateContent map[string]string) {
	genericpage := genericPageBuilder(cp)
	web.Get(url, GenerateHTMLwithTemplate(genericpage, templateContent))
	web.Get(cp.GeneratedCSSurl, GenerateCSS(genericpage))
	web.Get(cssurl, GenerateMenuCSS(cp.StretchBackground, cs))
}

// Wrap a lonely string in an entire webpage
func (cp *ContentPage) Surround(s string, tp map[string]string) (string, string) {
	cp.ContentHTML = s
	archpage := genericPageBuilder(cp)
	// NOTE: Use GetXML(true) instead of .String() or .GetHTML() because some things are rendered
	// differently with different text layout!
	return mustache.Render(archpage.GetXML(true), tp), archpage.GetCSS()
}

// Uses a given SimpleWebHandle as the contents for the the ContentPage contents
func (cp *ContentPage) WrapSimpleWebHandle(swh SimpleWebHandle, tp map[string]string) SimpleWebHandle {
	return func(val string) string {
		html, css := cp.Surround(swh(val), tp)
		web.Get(cp.GeneratedCSSurl, css)
		return html
	}
}

// Uses a given WebHandle as the contents for the the ContentPage contents
func (cp *ContentPage) WrapWebHandle(wh WebHandle, tp map[string]string) WebHandle {
	return func(ctx *web.Context, val string) string {
		html, css := cp.Surround(wh(ctx, val), tp)
		web.Get(cp.GeneratedCSSurl, css)
		return html
	}
}

// Uses a given SimpleContextHandle as the contents for the the ContentPage contents
func (cp *ContentPage) WrapSimpleContextHandle(sch SimpleContextHandle, tp map[string]string) SimpleContextHandle {
	return func(ctx *web.Context) string {
		html, css := cp.Surround(sch(ctx), tp)
		web.Get(cp.GeneratedCSSurl, css)
		return html
	}
}

func InitSystem() *ConnectionPool {
	// These common ones are missing!
	mime.AddExtensionType(".txt", "text/plain; charset=utf-8")
	mime.AddExtensionType(".ico", "image/x-icon")

	// Create a Redis connection pool
	return NewRedisConnectionPool()

	//if err != nil {
	//	panic("ERROR: Can't connect to redis")
	//}
	//defer pool.Close()

	// The login system, returns a *UserState
	//return InitUserSystem(pool)
}
