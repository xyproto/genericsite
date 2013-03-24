package genericsite

// OK, only user-related stuff, 23-03-13

import (
	"crypto/sha256"
	"errors"
	"io"
	"math/rand"
	"strings"
	"time"

	. "github.com/xyproto/browserspeak"
	"github.com/xyproto/web"
)

type UserState struct {
	// see: http://redis.io/topics/data-types
	users       *RedisHashMap   // Hash map of users, with several different fields per user ("loggedin", "confirmed", "email" etc)
	usernames   *RedisSet       // A list of all usernames, for easy enumeration
	unconfirmed *RedisSet       // A list of unconfirmed usernames, for easy enumeration
	pool        *ConnectionPool // A connection pool for Redis
}

// An Engine is a specific piece of a website
// This part handles the login/logout/registration/confirmation pages

const (
	ONLY_LOGIN      = "100"
	ONLY_LOGOUT     = "010"
	ONLY_REGISTER   = "001"
	EXCEPT_LOGIN    = "011"
	EXCEPT_LOGOUT   = "101"
	EXCEPT_REGISTER = "110"
	NOTHING         = "000"

	MINIMUM_CONFIRMATION_CODE_LENGTH = 20
	USERNAME_ALLOWED_LETTERS         = "abcdefghijklmnopqrstuvwxyzæøåABCDEFGHIJKLMNOPQRSTUVWXYZÆØÅ_0123456789"
)

type UserEngine struct {
	state *UserState
}

func NewUserEngine(pool *ConnectionPool) *UserEngine {
	var ue UserEngine
	ue.Init(pool)
	return &ue
}

func (state *UserState) GetPool() *ConnectionPool {
	return state.pool
}

func (ue *UserEngine) GetState() *UserState {
	return ue.state
}

func (ue UserEngine) Init(pool *ConnectionPool) {

	// For the secure cookies
	// This must happen before the random seeding, or 
	// else people will have to log in again after every server restart
	web.Config.CookieSecret = RandomCookieFriendlyString(30)

	rand.Seed(time.Now().UnixNano())

	ue.state = createUserState(pool)
}

// Checks if the current user is logged in as a user right now
func (state *UserState) UserRights(ctx *web.Context) bool {
	if username := GetBrowserUsername(ctx); username != "" {
		return state.IsLoggedIn(username)
	}
	return false
}

// TODO: Consider changing ShowMenu in the Engine interface
func (ue *UserEngine) ShowMenu(url string, ctx *web.Context) bool {
	return true
}

// The login menu
func (ue *UserEngine) ShowLoginMenu(url string, ctx *web.Context) bool {
	if url == "/login" {
		return false
	}
	if ue.state.UserRights(ctx) {
		return false
	}
	return true
}

// The logout menu
func (ue *UserEngine) ShowLogoutMenu(url string, ctx *web.Context) bool {
	if url == "/logout" {
		return false
	}
	if ue.state.UserRights(ctx) {
		return true
	}
	return false
}

// The register menu
func (ue *UserEngine) ShowRegisterMenu(url string, ctx *web.Context) bool {
	if url == "/register" {
		return false
	}
	if ue.state.UserRights(ctx) {
		return false
	}
	return true
}

// TODO: Don't return false if there is an error, the user may exist
func (state *UserState) HasUser(username string) bool {
	val, err := state.usernames.Has(username)
	if err != nil {
		// This happened at concurrent connections before introducing the connection pool
		panic("ERROR: Lost connection to Redis?")
	}
	return val
}

// Creates a user without doing ANY checks
func AddUserUnchecked(state *UserState, username, password, email string) {
	// Add the user
	state.usernames.Add(username)

	// Add password and email
	state.users.Set(username, "password", password)
	state.users.Set(username, "email", email)

	// Addditional fields
	additionalfields := []string{"loggedin", "confirmed", "admin"}
	for _, fieldname := range additionalfields {
		state.users.Set(username, fieldname, "false")
	}
}

func (state *UserState) IsConfirmed(username string) bool {
	hasUser := state.HasUser(username)
	if !hasUser {
		return false
	}
	confirmed, err := state.users.Get(username, "confirmed")
	if err != nil {
		return false
	}
	return TruthValue(confirmed)
}

func CorrectPassword(state *UserState, username, password string) bool {
	hashedPassword, err := state.users.Get(username, "password")
	if err != nil {
		return false
	}
	if hashedPassword == HashPasswordVersion2(password) {
		return true
	}
	//if hashedPassword == HashPasswordVersion1(password) {
	//	return true
	//}
	return false
}

func (state *UserState) GetConfirmationSecret(username string) string {
	secret, err := state.users.Get(username, "secret")
	if err != nil {
		return ""
	}
	return secret
}

// Goes through all the secrets of all the unconfirmed users
// and checks if this secret already is in use
func AlreadyHasSecret(state *UserState, secret string) bool {
	unconfirmedUsernames, err := state.unconfirmed.GetAll()
	if err != nil {
		return false
	}
	for _, aUsername := range unconfirmedUsernames {
		aSecret, err := state.users.Get(aUsername, "secret")
		if err != nil {
			// TODO: Inconsistent user, log this
			continue
		}
		if secret == aSecret {
			// Found it
			return true
		}
	}
	return false
}

// Create a user by adding the username to the list of usernames
func GenerateConfirmUser(state *UserState) WebHandle {
	return func(ctx *web.Context, val string) string {
		secret := val

		unconfirmedUsernames, err := state.unconfirmed.GetAll()
		if err != nil {
			return MessageOKurl("Confirmation", "All users are confirmed already.", "/register")
		}

		// TODO: Only generate unique secrets

		// Find the username by looking up the secret on unconfirmed users
		username := ""
		for _, aUsername := range unconfirmedUsernames {
			aSecret, err := state.users.Get(aUsername, "secret")
			if err != nil {
				// TODO: Inconsistent user! Log this.
				continue
			}
			if secret == aSecret {
				// Found the right user
				username = aUsername
				break
			}
		}

		// Check that the user is there
		if username == "" {
			// Say "no longer" because we don't care about people that just try random confirmation links
			return MessageOKurl("Confirmation", "The confirmation link is no longer valid.", "/register")
		}
		hasUser := state.HasUser(username)
		if !hasUser {
			return MessageOKurl("Confirmation", "The user you wish to confirm does not exist anymore.", "/register")
		}

		// Remove from the list of unconfirmed usernames
		state.unconfirmed.Del(username)
		// Remove the secret from the user
		state.users.Del(username, "secret")

		// Mark user as confirmed
		state.users.Set(username, "confirmed", "true")

		return MessageOKurl("Confirmation", "Thank you "+username+", you can now log in.", "/login")
	}
}

// Log in a user by changing the loggedin value
func GenerateLoginUser(state *UserState) WebHandle {
	return func(ctx *web.Context, val string) string {
		// Fetch password from ctx
		password, found := ctx.Params["password"]
		if !found {
			return MessageOKback("Login", "Can't log in without a password.")
		}
		username := val
		if username == "" {
			return MessageOKback("Login", "Can't log in with a blank username.")
		}
		if !state.HasUser(username) {
			return MessageOKback("Login", "User "+username+" does not exist, could not log in.")
		}
		if !state.IsConfirmed(username) {
			return MessageOKback("Login", "The email for "+username+" has not been confirmed, check your email and follow the link.")
		}

		// TODO: Hash password, check with hash from database

		if !CorrectPassword(state, username, password) {
			return MessageOKback("Login", "Wrong password.")
		}

		// Log in the user by changing the database and setting a secure cookie
		state.users.Set(username, "loggedin", "true")
		state.SetBrowserUsername(ctx, username)

		// TODO: Use a welcoming messageOK where the user can see when he/she last logged in and from which host

		// TODO: Then redirect to the page the user was at before logging in
		if username == "admin" {
			ctx.SetHeader("Refresh", "0; url=/admin", true)
		} else {
			ctx.SetHeader("Refresh", "0; url=/", true)
		}

		return ""
	}
}

func HashPasswordVersion2(password string) string {
	hasher := sha256.New()
	// TODO: Read up on password hashing
	io.WriteString(hasher, password+"some salt is better than none")
	return string(hasher.Sum(nil))
}

// TODO: Forgot username? Enter email, send username.
// TODO: Lost confirmation link? Enter mail, Receive confirmation link.
// TODO: Forgot password? Enter mail, receive reset-password link.
// TODO: Make sure not two usernames can register at once before confirming
// TODO: Only one username per email address? (meh? can use more than one address?=
// TODO: Maximum 1 confirmation email per email adress
// TODO: Maximum 1 forgot username per email adress per day
// TODO: Maximum 1 forgot password per email adress per day
// TODO: Maximum 1 lost confirmation link per email adress per day
// TODO: Link for "Did you not request this email? Click here" i alle eposter som sendes.
// TODO: Rate limiting, maximum rate per minute or day

// Register a new user
func GenerateRegisterUser(state *UserState) WebHandle {
	return func(ctx *web.Context, val string) string {

		// Password checks
		password1, found := ctx.Params["password1"]
		if password1 == "" || !found {
			return MessageOKback("Register", "Can't register without a password.")
		}
		password2, found := ctx.Params["password2"]
		if password2 == "" || !found {
			return MessageOKback("Register", "Please confirm the password by typing it in twice.")
		}
		if password1 != password2 {
			return MessageOKback("Register", "The password and confirmation password must be equal.")
		}

		// Email checks
		email, found := ctx.Params["email"]
		if !found {
			return MessageOKback("Register", "Can't register without an email address.")
		}
		// must have @ and ., but no " "
		if !strings.Contains(email, "@") || !strings.Contains(email, ".") || strings.Contains(email, " ") {
			return MessageOKback("Register", "Please use a valid email address.")
		}

		// Username checks
		username := val
		if username == "" {
			return MessageOKback("Register", "Can't register without a username.")
		}
		if state.HasUser(username) {
			return MessageOKback("Register", "That user already exists, try another username.")
		}
		// Only some letters are allowed
	NEXT:
		for _, letter := range username {
			for _, allowedLetter := range USERNAME_ALLOWED_LETTERS {
				if letter == allowedLetter {
					continue NEXT
				}
			}
			return MessageOKback("Register", "Only a-å, A-Å, 0-9 and _ are allowed in usernames.")
		}
		if username == password1 {
			return MessageOKback("Register", "Username and password must be different, try another password.")
		}

		adminuser := false
		// A special case
		if username == "admin" {
			// The first user to register with the username "admin" becomes the administrator
			adminuser = true
		}

		// Register the user
		password := HashPasswordVersion2(password1)
		AddUserUnchecked(state, username, password, email)

		// Mark user as administrator if that is the case
		if adminuser {
			// This does not set the username to admin,
			// but sets the admin field to true
			state.users.Set(username, "admin", "true")
		}

		// The confirmation code must be a minimum of 8 letters long
		length := MINIMUM_CONFIRMATION_CODE_LENGTH
		secretConfirmationCode := RandomHumanFriendlyString(length)
		for AlreadyHasSecret(state, secretConfirmationCode) {
			// Increase the length of the secret random string every time there is a collision
			length++
			secretConfirmationCode = RandomHumanFriendlyString(length)
			if length > 100 {
				// Something is seriously wrong if this happens
				// TODO: Log this and sysexit
			}
		}

		// Send confirmation email
		ConfirmationEmail("archlinux.no", "https://archlinux.no/confirm/"+secretConfirmationCode, username, email)

		// Register the need to be confirmed
		state.unconfirmed.Add(username)
		state.users.Set(username, "secret", secretConfirmationCode)

		// Redirect
		//ctx.SetHeader("Refresh", "0; url=/login", true)

		return MessageOKurl("Registration complete", "Thanks for registering, the confirmation e-mail has been sent.", "/login")
	}
}

// Log out a user by changing the loggedin value
func GenerateLogoutCurrentUser(state *UserState) SimpleContextHandle {
	return func(ctx *web.Context) string {
		username := GetBrowserUsername(ctx)
		if username == "" {
			return MessageOKback("Logout", "No user to log out")
		}
		if !state.HasUser(username) {
			return MessageOKback("Logout", "user "+username+" does not exist, could not log out")
		}

		// TODO: Check if the user is logged in already

		// Log out the user by changing the database, the cookie can stay
		state.users.Set(username, "loggedin", "false")

		//return "OK, user " + username + " logged out"

		// TODO: Redirect to the page the user was at before logging out
		//ctx.SetHeader("Refresh", "0; url=/", true)

		//return ""

		return MessageOKurl("Logout", username+" is now logged out. Hope to see you soon!", "/login")
	}
}

// Checks if the given username is logged in or not
func (state *UserState) IsLoggedIn(username string) bool {
	if !state.HasUser(username) {
		return false
	}
	status, err := state.users.Get(username, "loggedin")
	if err != nil {
		return false
	}
	return TruthValue(status)
}

// Gets the username that is stored in a cookie in the browser, if available
func GetBrowserUsername(ctx *web.Context) string {
	username, _ := ctx.GetSecureCookie("user")
	return username
}

func (state *UserState) SetBrowserUsername(ctx *web.Context, username string) error {
	if username == "" {
		return errors.New("Can't set cookie for empty username")
	}
	if !state.HasUser(username) {
		return errors.New("Can't store cookie for non-existsing user")
	}
	// TODO: Users should be able to select their own cookie timeout
	// Create a cookie that lasts for one hour,
	// this is the equivivalent of a session for a given username
	ctx.SetSecureCookiePath("user", username, 3600, "/")
	//"Cookie stored: user = " + username + "."
	return nil
}

func GenerateNoJavascriptMessage() SimpleContextHandle {
	return func(ctx *web.Context) string {
		return MessageOKback("JavaScript error", "Logging in without cookies and javascript enabled, in a modern browser, is not yet supported.<br />Elinks will be supported in the future.")
	}
}

func createUserState(pool *ConnectionPool) *UserState {
	// For the database
	state := new(UserState)
	state.users = NewRedisHashMap(pool, "users")
	state.usernames = NewRedisSet(pool, "usernames")
	state.unconfirmed = NewRedisSet(pool, "unconfirmed")
	state.pool = pool
	return state
}

func LoginCP(basecp BaseCP, state *UserState, url string) *ContentPage {
	cp := basecp(state)
	cp.ContentTitle = "Login"
	cp.ContentHTML = LoginForm()
	cp.ContentJS += OnClick("#loginButton", "$('#loginForm').get(0).setAttribute('action', '/login/' + $('#username').val());")

	// Hide the Login menu if we're on the Login page
	// TODO: Replace with the entire Javascript expression, not just menuNop?
	//cp.HeaderJS = strings.Replace(cp.HeaderJS, "menuLogin", "menuNop", 1)
	//cp.ContentJS += Hide("#menuLogin")

	cp.Url = url
	return cp
}

func RegisterCP(basecp BaseCP, state *UserState, url string) *ContentPage {
	cp := basecp(state)
	cp.ContentTitle = "Register"
	cp.ContentHTML = RegisterForm()
	cp.ContentJS += OnClick("#registerButton", "$('#registerForm').get(0).setAttribute('action', '/register/' + $('#username').val());")
	cp.Url = url

	// Hide the Register menu if we're on the Register page
	// TODO: Replace with the entire Javascript expression, not just menuNop?
	//cp.HeaderJS = strings.Replace(cp.HeaderJS, "menuRegister", "menuNop", 1)
	//cp.ContentJS += Hide("#menuRegister")

	return cp
}

func (ue *UserEngine) ServeSystem() {
	state := ue.state
	web.Post("/register/(.*)", GenerateRegisterUser(state))
	web.Post("/login/(.*)", GenerateLoginUser(state))
	web.Post("/login", GenerateNoJavascriptMessage())
	web.Get("/logout", GenerateLogoutCurrentUser(state))
	web.Get("/confirm/(.*)", GenerateConfirmUser(state))
}
