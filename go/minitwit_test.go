package main

import (
	"fmt"
	"io"
	mathRand "math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const BASE_URL = "http://localhost:5000"

func do_register(username string, password string, password2 string, email string) (*http.Response, error) {
	// Helper function to register a user
	if password2 == "" {
		password2 = password
	}
	if email == "" {
		email = username + "@example.com"
	}
	jar, _ := cookiejar.New(nil)
	http_session := &http.Client{Jar: jar}
	return http_session.PostForm(BASE_URL+"/register", url.Values{
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

func do_add_message(http_session *http.Client, text string, t *testing.T) (*http.Response, error) {
	// Records a message
	r, err := http_session.PostForm(BASE_URL+"/add_message", url.Values{"text": {text}})
	if text != "" {
		assert.True(t, resp_contains("Your message was recorded", r))
	}
	return r, err
}

func resp_to_text(r *http.Response) string {
	// Helper function not present in original test script
	defer r.Body.Close()
	r_text, _ := io.ReadAll(r.Body)
	return string(r_text)
}

func resp_contains(substring string, r *http.Response) bool {
	// Helper function not present in original test script
	return strings.Contains(resp_to_text(r), substring)
}

func unique_user() string {
	// Helper function not present in original test script
	return fmt.Sprintf("%s%d", "user", mathRand.Int())
}

func Test_register(t *testing.T) {
	// Make sure registering works
	user1 := unique_user()
	r, _ := do_register(user1, "default", "", "")
	assert.True(t, resp_contains("You were successfully registered and can login now", r))
	r, _ = do_register(user1, "default", "", "")
	assert.True(t, resp_contains("The username is already taken", r))
	r, _ = do_register("", "default", "", "")
	assert.True(t, resp_contains("You have to enter a username", r))
	meh := unique_user()
	r, _ = do_register(meh, "", "", "")
	assert.True(t, resp_contains("You have to enter a password", r))
	r, _ = do_register(meh, "x", "y", "")
	assert.True(t, resp_contains("The two passwords do not match", r))
	r, _ = do_register(meh, "foo", "", "broken")
	assert.True(t, resp_contains("You have to enter a valid email address", r))
}

func Test_login_logout(t *testing.T) {
	user1 := unique_user()
	user2 := unique_user()
	r, http_session, _ := do_register_and_login(user1, "default")
	assert.True(t, resp_contains("You were logged in", r))
	r, _ = do_logout(http_session)
	assert.True(t, resp_contains("You were logged out", r))
	r, _, _ = do_login(user1, "wrongpassword")
	assert.True(t, resp_contains("Invalid password", r))
	r, _, _ = do_login(user2, "wrongpassword")
	assert.True(t, resp_contains("Invalid username", r))
}

func Test_message_recording(t *testing.T) {
	// Check if adding messages works
	foo := unique_user()
	_, http_session, _ := do_register_and_login(foo, "default")
	do_add_message(http_session, "test message 1", t)
	do_add_message(http_session, "<test message 2>", t)
	r, _ := http.Get(BASE_URL + "/public")
	r_text := resp_to_text(r)
	assert.True(t, strings.Contains(r_text, "test message 1"))
	assert.True(t, strings.Contains(r_text, "&lt;test message 2&gt;"))
}

func Test_timelines(t *testing.T) {
	// Make sure that timelines work
	foo := unique_user()
	_, http_session, _ := do_register_and_login(foo, "default")
	do_add_message(http_session, "the message by "+foo, t)
	do_logout(http_session)
	bar := unique_user()
	_, http_session, _ = do_register_and_login(bar, "default")
	do_add_message(http_session, "the message by "+bar, t)
	r, _ := http_session.Get(BASE_URL + "/public")
	r_text := resp_to_text(r)
	assert.True(t, strings.Contains(r_text, "the message by "+foo))
	assert.True(t, strings.Contains(r_text, "the message by "+bar))

	// bar's timeline should just show bar's message
	r, _ = http_session.Get(BASE_URL + "/")
	r_text = resp_to_text(r)
	assert.True(t, !strings.Contains(r_text, "the message by "+foo))
	assert.True(t, strings.Contains(r_text, "the message by "+bar))

	// now let's follow foo
	r, _ = http_session.Get(BASE_URL + "/" + foo + "/follow")
	assert.True(t, resp_contains("You are now following &#34;"+foo+"&#34;", r))

	// we should now see foo's message
	r, _ = http_session.Get(BASE_URL + "/")
	r_text = resp_to_text(r)
	assert.True(t, strings.Contains(r_text, "the message by "+foo))
	assert.True(t, strings.Contains(r_text, "the message by "+bar))

	// but on the user's page we only want the user's message
	r, _ = http_session.Get(BASE_URL + "/" + bar)
	r_text = resp_to_text(r)
	assert.True(t, !strings.Contains(r_text, "the message by "+foo))
	assert.True(t, strings.Contains(r_text, "the message by "+bar))
	r, _ = http_session.Get(BASE_URL + "/" + foo)
	r_text = resp_to_text(r)
	assert.True(t, strings.Contains(r_text, "the message by "+foo))
	assert.True(t, !strings.Contains(r_text, "the message by "+bar))

	// now unfollow and check if that worked
	r, _ = http_session.Get(BASE_URL + "/" + foo + "/unfollow")
	assert.True(t, resp_contains("You are no longer following &#34;"+foo+"&#34;", r))
	r, _ = http_session.Get(BASE_URL + "/")
	r_text = resp_to_text(r)
	assert.True(t, !strings.Contains(r_text, "the message by "+foo))
	assert.True(t, strings.Contains(r_text, "the message by "+bar))
}
