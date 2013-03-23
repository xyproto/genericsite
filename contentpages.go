package genericsite

// TODO: Move some of these to "browserspeak"

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

type UserState struct {
	// see: http://redis.io/topics/data-types
	users       *RedisHashMap   // Hash map of users, with several different fields per user ("loggedin", "confirmed", "email" etc)
	usernames   *RedisSet       // A list of all usernames, for easy enumeration
	unconfirmed *RedisSet       // A list of unconfirmed usernames, for easy enumeration
	pool        *ConnectionPool // A connection pool for Redis
}

type ColorScheme struct {
	darkgray           string
	nicecolor          string
	menu_link          string
	menu_hover         string
	menu_active        string
	default_background string
}

const (
	JQUERY_VERSION = "1.9.1"
)

// browserspeak
var globalStringCache map[string]string

// browserspeak
// Wrap a SimpleContextHandle so that the output is cached (with an id)
// Do not cache functions with side-effects! (that sets the mimetype for instance)
// The safest thing for now is to only cache images.
func CacheWrapper(id string, f SimpleContextHandle) SimpleContextHandle {
	return func(ctx *web.Context) string {
		if _, ok := globalStringCache[id]; !ok {
			globalStringCache[id] = f(ctx)
		}
		return globalStringCache[id]
	}
}

// browserspeak
func Publish(url, filename string, cache bool) {
	if cache {
		web.Get(url, CacheWrapper(url, File(filename)))
	} else {
		web.Get(url, File(filename))
	}
}

// TODO: get style values from a file instead?
func AddHeader(page *Page, js string) {
	AddGoogleFonts(page, []string{"Armata"}) //, "Junge"})
	// TODO: Move to browserspeak
	page.MetaCharset("UTF-8")
	AddScriptToHeader(page, js)
}

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

	// Javascript that applies everywhere
	//cp.contentJS += HideIfNot("/showmenu/login", "#menuLogin")
	//cp.contentJS += HideIfNot("/showmenu/logout", "#menuLogout")
	//cp.contentJS += HideIfNot("/showmenu/register", "#menuRegister")
	//cp.contentJS += HideIfNotLoginLogoutRegister("/showmenu/loginlogoutregister", "#menuLogin", "#menuLogout", "#menuRegister")
	//cp.contentJS += ShowIfLoginLogoutRegister("/showmenu/loginlogoutregister", "#menuLogin", "#menuLogout", "#menuRegister")

	// This only works at first page load in Internet Explorer 8. Fun times. Oh well, why bother.
	cp.HeaderJS += ShowIfLoginLogoutRegister("/showmenu/loginlogoutregister", "#menuLogin", "#menuLogout", "#menuRegister")

	// This in combination with hiding the link in genericsite.go is cool, but the layout becomes weird :/
	//cp.headerJS += ShowAnimatedIf("/showmenu/admin", "#menuAdmin")

	// This keeps the layout but is less cool
	cp.HeaderJS += HideIfNot("/showmenu/admin", "#menuAdmin")

	cp.Url = "/" // To be filled in when published

	// The default color scheme
	var cs ColorScheme
	cs.darkgray = "#202020"
	cs.nicecolor = "#5080D0"   // nice blue
	cs.menu_link = "#c0c0c0"   // light gray
	cs.menu_hover = "#efefe0"  // light gray, somewhat yellow
	cs.menu_active = "#ffffff" // white
	cs.default_background = "#000030"
	cp.colorScheme = &cs

	return &cp
}

// TODO: Consider using Mustache for replacing elements after the page has been generated
// (for showing/hiding "login", "logout" or "register"
func genericPageBuilder(cp *ContentPage) *Page {
	// TODO: Record the time from one step out, because content may be generated and inserted into this generated conten
	startTime := time.Now()

	page := NewHTML5Page(cp.title + " " + cp.subtitle)

	page.LinkToCSS(cp.generatedCSSurl)
	for _, cssurl := range cp.extraCSSurls {
		page.LinkToCSS(cssurl)
	}
	page.LinkToJS(cp.jqueryJSurl)
	page.LinkToFavicon(cp.faviconurl)

	AddHeader(page, cp.headerJS)
	AddBodyStyle(page, cp.bgImageURL, cp.stretchBackground)
	AddTopBox(page, cp.title, cp.subtitle, cp.searchURL, cp.searchButtonText, cp.backgroundTextureURL, cp.roundedLook, cp.colorScheme)

	// TODO:
	// Use something dynamic to add or remove /login and /register depending on the login status
	// The login status can be fetched over AJAX or REST or something.

	// TODO: Move the menubox into the TopBox

	// TODO: Do this with templates instead
	// Hide login/logout/register by default
	hidelist := []string{"/login", "/logout", "/register"} //, "/admin"}
	AddMenuBox(page, cp.links, hidelist, cp.darkBackgroundTextureURL)

	AddContent(page, cp.contentTitle, cp.contentHTML+BodyJS(cp.contentJS))

	elapsed := time.Since(startTime)
	AddFooter(page, cp.footerText, cp.footerTextColor, cp.footerColor, elapsed)

	return page
}

// Publish a list of ContentPaages, a colorscheme and template content
func PublishCPs(pc PageCollection, cs *ColorScheme, tp map[string]string, cssurl string) {
	// For each content page in the page collection
	for _, cp := range pc {
		// TODO: different css urls for all of these?
		cp.Pub(cp.url, cssurl, cs, tp)
	}
}

type BaseCP func(state *UserState) *ContentPage

func ServeSite(basecp BaseCP, userState *UserState, cps PageCollection, tp map[string]string) {
	// Add pages for login, logout and register
	cps = append(cps, *LoginCP(basecp, userState, "/login"))
	cps = append(cps, *RegisterCP(basecp, userState, "/register"))

	cs := basecp(userState).colorScheme
	PublishCPs(cps, cs, tp, "/css/extra.css")

	ServeSearchPages(basecp, userState, cps, cs, tp)
	ServeAdminPages(basecp, userState, cs, tp)

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
a:link {color:` + cs.menu_link + `;}
a:visited {color:` + cs.menu_link + `;}
a:hover {color:` + cs.menu_hover + `;}
a:active {color:` + cs.menu_active + `;}
`
		// The load order of background-color, background-size and background-image
		// is actually significant in Chrome! Do not reorder lightly!
		if stretchBackground {
			retval = "body {\nbackground-color: " + cs.default_background + ";\nbackground-size: cover;\n}\n" + retval
		} else {
			retval = "body {\nbackground-color: " + cs.default_background + ";\n}\n" + retval
		}
		return retval
	}
}

// Make an html and css page available
func (cp *ContentPage) Pub(url, cssurl string, cs *ColorScheme, templateContent map[string]string) {
	genericpage := genericPageBuilder(cp)
	web.Get(url, GenerateHTMLwithTemplate(genericpage, templateContent))
	web.Get(cp.generatedCSSurl, GenerateCSS(genericpage))
	web.Get(cssurl, GenerateMenuCSS(cp.stretchBackground, cs))
}

// Wrap a lonely string in an entire webpage
func (cp *ContentPage) Surround(s string, tp map[string]string) (string, string) {
	cp.contentHTML = s
	archpage := genericPageBuilder(cp)
	// NOTE: Use GetXML(true) instead of .String() or .GetHTML() because some things are rendered
	// differently with different text layout!
	return mustache.Render(archpage.GetXML(true), tp), archpage.GetCSS()
}

// Uses a given SimpleWebHandle as the contents for the the ContentPage contents
func (cp *ContentPage) WrapSimpleWebHandle(swh SimpleWebHandle, tp map[string]string) SimpleWebHandle {
	return func(val string) string {
		html, css := cp.Surround(swh(val), tp)
		web.Get(cp.generatedCSSurl, css)
		return html
	}
}

// Uses a given WebHandle as the contents for the the ContentPage contents
func (cp *ContentPage) WrapWebHandle(wh WebHandle, tp map[string]string) WebHandle {
	return func(ctx *web.Context, val string) string {
		html, css := cp.Surround(wh(ctx, val), tp)
		web.Get(cp.generatedCSSurl, css)
		return html
	}
}

// Uses a given SimpleContextHandle as the contents for the the ContentPage contents
func (cp *ContentPage) WrapSimpleContextHandle(sch SimpleContextHandle, tp map[string]string) SimpleContextHandle {
	return func(ctx *web.Context) string {
		html, css := cp.Surround(sch(ctx), tp)
		web.Get(cp.generatedCSSurl, css)
		return html
	}
}

func InitSystem() *UserState {
	// These common ones are missing!
	mime.AddExtensionType(".txt", "text/plain; charset=utf-8")
	mime.AddExtensionType(".ico", "image/x-icon")

	// Create a Redis connection pool
	pool := NewRedisConnectionPool()
	//if err != nil {
	//	panic("ERROR: Can't connect to redis")
	//}
	defer pool.Close()

	// The login system, returns a *UserState
	return InitUserSystem(pool)
}
