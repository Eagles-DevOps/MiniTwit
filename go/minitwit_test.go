package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const BASE_URL = "http://server:15000"

func do_register(username string, password string, password2 string, email string) (*http.Response, error) {
	// Helper function to register a user
	if password2 == "" {
		password2 = password
	}
	if email == "" {
		email = username + "@example.com"
	}
	return http.PostForm(BASE_URL+"/register", url.Values{
		"username":  {username},
		"password":  {password},
		"password2": {password2},
		"email":     {email},
	})
}

func do_login(username string, password string) (*http.Response, *http.Client, error) {
	// Helper function to login
	jar, _ := cookiejar.New(nil)
	http_session := &http.Client{Jar: jar}
	r, http_err := http_session.PostForm(BASE_URL+"/login", url.Values{
		"username": {username},
		"password": {password},
	})
	return r, http_session, http_err
}

func do_register_and_login(username string, password string) (*http.Response, *http.Client, error) {
	// Registers and logs in in one go
	do_register(username, password, "", "")
	return do_login(username, password)
}

func do_logout(http_session *http.Client) (*http.Response, error) {
	// Helper function to logout
	return http_session.Get(BASE_URL + "/logout") // Follows redirects by default
}

func txt_in_resp(substring string, r *http.Response) bool {
	// Helper function not present in original test script
	defer r.Body.Close()
	r_text, _ := io.ReadAll(r.Body)
	r_text_str := string(r_text)
	return strings.Contains(r_text_str, substring)
}

func do_add_message(http_session *http.Client, text string, t *testing.T) (*http.Response, error) {
	// Records a message
	r, err := http_session.PostForm(BASE_URL+"/add_message", url.Values{"text": {text}})
	if text != "" {
		assert.True(t, txt_in_resp("Your message was recorded", r))
	}
	return r, err
}

func Test_register(t *testing.T) {
	// Make sure registering works
	r, _ := do_register("user1", "default", "", "")
	assert.True(t, txt_in_resp("You were successfully registered and can login now", r))
	r, _ = do_register("user1", "default", "", "")
	assert.True(t, txt_in_resp("The username is already taken", r))
	r, _ = do_register("", "default", "", "")
	assert.True(t, txt_in_resp("You have to enter a username", r))
	r, _ = do_register("meh", "", "", "")
	assert.True(t, txt_in_resp("You have to enter a password", r))
	r, _ = do_register("meh", "x", "y", "")
	assert.True(t, txt_in_resp("The two passwords do not match", r))
	r, _ = do_register("meh", "foo", "", "broken")
	assert.True(t, txt_in_resp("You have to enter a valid email address", r))
}

func Test_login_logout(t *testing.T) {
	r, http_session, _ := do_register_and_login("user1", "default")
	assert.True(t, txt_in_resp("You were logged in", r))
	r, _ = do_logout(http_session)
	assert.True(t, txt_in_resp("You were logged out", r))
	r, _, _ = do_login("user1", "wrongpassword")
	assert.True(t, txt_in_resp("Invalid password", r))
	r, _, _ = do_login("user2", "wrongpassword")
	assert.True(t, txt_in_resp("Invalid username", r))
}

func Test_message_recording(t *testing.T) {
	// Check if adding messages works
	_, http_session, _ := do_register_and_login("foo", "default")
	do_add_message(http_session, "test message 1", t)
	do_add_message(http_session, "<test message 2>", t)
	r, _ := http.Get(BASE_URL + "/public")
	assert.True(t, txt_in_resp("test message 1", r))
	assert.True(t, txt_in_resp("&lt;test message 2&gt;", r))
}

func Test_timelines(t *testing.T) {
	// Make sure that timelines work
	_, http_session, _ := do_register_and_login("foo", "default")
	do_add_message(http_session, "the message by foo", t)
	do_logout(http_session)
	_, http_session, _ = do_register_and_login("bar", "default")
	do_add_message(http_session, "the message by bar", t)
	r, _ := http_session.Get(BASE_URL + "/public")
	assert.True(t, txt_in_resp("the message by foo", r))
	assert.True(t, txt_in_resp("the message by bar", r))

	// bar's timeline should just show bar's message
	r, _ = http_session.Get(BASE_URL + "/")
	assert.True(t, !txt_in_resp("the message by foo", r))
	assert.True(t, txt_in_resp("the message by bar", r))

	// now let's follow foo
	r, _ = http_session.Get(BASE_URL + "/foo/follow")
	assert.True(t, txt_in_resp("You are now following &#34;foo&#34;", r))

	// we should now see foo's message
	r, _ = http_session.Get(BASE_URL + "/")
	assert.True(t, txt_in_resp("the message by foo", r))
	assert.True(t, txt_in_resp("the message by bar", r))

	// but on the user's page we only want the user's message
	r, _ = http_session.Get(BASE_URL + "/bar")
	assert.True(t, !txt_in_resp("the message by foo", r))
	assert.True(t, txt_in_resp("the message by bar", r))
	r, _ = http_session.Get(BASE_URL + "/foo")
	assert.True(t, txt_in_resp("the message by foo", r))
	assert.True(t, !txt_in_resp("the message by bar", r))

	// now unfollow and check if that worked
	r, _ = http_session.Get(BASE_URL + "/foo/unfollow")
	assert.True(t, txt_in_resp("You are no longer following &#34;foo&#34;", r))
	r, _ = http_session.Get(BASE_URL + "/")
	assert.True(t, !txt_in_resp("the message by foo", r))
	assert.True(t, txt_in_resp("the message by bar", r))
}
