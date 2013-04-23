package genericsite

import (
	"errors"

	"github.com/xyproto/simpleredis"
	"github.com/xyproto/web"
)

const (
	LOGINCOOKIETIME = 3600 * 24 // The login cookie should last 24 hours
)

type UserState struct {
	// see: http://redis.io/topics/data-types
	users       *simpleredis.HashMap        // Hash map of users, with several different fields per user ("loggedin", "confirmed", "email" etc)
	usernames   *simpleredis.Set            // A list of all usernames, for easy enumeration
	unconfirmed *simpleredis.Set            // A list of unconfirmed usernames, for easy enumeration
	pool        *simpleredis.ConnectionPool // A connection pool for Redis
}

// Also creates a new ConnectionPool
func NewUserState() *UserState {
	pool := simpleredis.NewConnectionPool()
	state := new(UserState)
	state.users = simpleredis.NewHashMap(pool, "users")
	state.usernames = simpleredis.NewSet(pool, "usernames")
	state.unconfirmed = simpleredis.NewSet(pool, "unconfirmed")
	state.pool = pool
	return state
}

func (state *UserState) GetPool() *simpleredis.ConnectionPool {
	return state.pool
}

func (state *UserState) Close() {
	state.pool.Close()
}

// Checks if the current user is logged in as a user right now
func (state *UserState) UserRights(ctx *web.Context) bool {
	if username := GetBrowserUsername(ctx); username != "" {
		return state.IsLoggedIn(username)
	}
	return false
}

func (state *UserState) HasUser(username string) bool {
	val, err := state.usernames.Has(username)
	if err != nil {
		// This happened at concurrent connections before introducing the connection pool
		panic("ERROR: Lost connection to Redis?")
	}
	return val
}

func (state *UserState) GetBooleanField(username, fieldname string) bool {
	hasUser := state.HasUser(username)
	if !hasUser {
		return false
	}
	chatting, err := state.users.Get(username, fieldname)
	if err != nil {
		return false
	}
	return TruthValue(chatting)
}

func (state *UserState) SetBooleanField(username, fieldname string, val bool) {
	strval := "false"
	if val {
		strval = "true"
	}
	state.users.Set(username, fieldname, strval)
}

func (state *UserState) IsConfirmed(username string) bool {
	return state.GetBooleanField(username, "confirmed")
}

// Checks if the given username is logged in or not
func (state *UserState) IsLoggedIn(username string) bool {
	if !state.HasUser(username) {
		return false
	}
	status, err := state.users.Get(username, "loggedin")
	if err != nil {
		// Returns "no" if the status can not be retrieved
		return false
	}
	return TruthValue(status)
}

// Checks if the current user is logged in as Administrator right now
func (state *UserState) AdminRights(ctx *web.Context) bool {
	if username := GetBrowserUsername(ctx); username != "" {
		return state.IsLoggedIn(username) && state.IsAdmin(username)
	}
	return false
}

// Checks if the given username is an administrator
func (state *UserState) IsAdmin(username string) bool {
	if !state.HasUser(username) {
		return false
	}
	status, err := state.users.Get(username, "admin")
	if err != nil {
		return false
	}
	return TruthValue(status)
}

// Gets the username that is stored in a cookie in the browser, if available
func GetBrowserUsername(ctx *web.Context) string {
	username, _ := ctx.GetSecureCookie("user")
	// TODO: Return err, then the calling function should notify the user that cookies are needed
	return username
}

func (state *UserState) SetBrowserUsername(ctx *web.Context, username string) error {
	if username == "" {
		return errors.New("Can't set cookie for empty username")
	}
	if !state.HasUser(username) {
		return errors.New("Can't store cookie for non-existsing user")
	}
	timeout := state.GetCookieTimeout(username)
	// Create a cookie that lasts for a while ("timeout" seconds),
	// this is the equivivalent of a session for a given username.
	ctx.SetSecureCookiePath("user", username, timeout, "/")
	return nil
}

func (state *UserState) GetAllUsernames() ([]string, error) {
	return state.usernames.GetAll()
}

func (state *UserState) GetEmail(username string) (string, error) {
	return state.users.Get(username, "email")
}

func (state *UserState) GetPasswordHash(username string) (string, error) {
	return state.users.Get(username, "password")
}

func (state *UserState) GetAllUnconfirmedUsernames() ([]string, error) {
	return state.unconfirmed.GetAll()
}

func (state *UserState) GetConfirmationCode(username string) (string, error) {
	return state.users.Get(username, "confirmationCode")
}

func (state *UserState) GetUsers() *simpleredis.HashMap {
	return state.users
}

// Add a user that has registered but not confirmed
func (state *UserState) AddUnconfirmed(username, confirmationCode string) {
	state.unconfirmed.Add(username)
	state.users.Set(username, "confirmationCode", confirmationCode)
}

// Remove a user that has registered but not confirmed
func (state *UserState) RemoveUnconfirmed(username string) {
	state.unconfirmed.Del(username)
	state.users.DelKey(username, "confirmationCode")
}

func (state *UserState) MarkConfirmed(username string) {
	state.users.Set(username, "confirmed", "true")
}

func (state *UserState) RemoveUser(username string) {
	state.usernames.Del(username)
	// Remove additional data as well
	state.users.DelKey(username, "loggedin")
}

func (state *UserState) SetAdminStatus(username string) {
	state.users.Set(username, "admin", "true")
}

func (state *UserState) RemoveAdminStatus(username string) {
	state.users.Set(username, "admin", "false")
}

// Creates a user without doing ANY checks
func (state *UserState) AddUserUnchecked(username, passwordHash, email string) {
	// Add the user
	state.usernames.Add(username)

	// Add password and email
	state.users.Set(username, "password", passwordHash)
	state.users.Set(username, "email", email)

	// Addditional fields
	additionalfields := []string{"loggedin", "confirmed", "admin"}
	for _, fieldname := range additionalfields {
		state.users.Set(username, fieldname, "false")
	}
}

func (state *UserState) SetLoggedIn(username string) {
	state.users.Set(username, "loggedin", "true")
}

func (state *UserState) SetLoggedOut(username string) {
	state.users.Set(username, "loggedin", "false")
}

// Get how long a login cookie should last
func (state *UserState) GetCookieTimeout(username string) int64 {
	// TODO: Store this in state.users
	return LOGINCOOKIETIME
}