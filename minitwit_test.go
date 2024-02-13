package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
)

const BASE_URL = "http://server:15000"

func register(username string, password string, password2 string, email string) (*http.Response, error) {
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

func login(username string, password string) (*http.Client, *http.Response, error) {
	// Helper function to login
	jar, _ := cookiejar.New(nil)
	http_session := &http.Client{Jar: jar}
	r, http_err := http_session.PostForm(BASE_URL+"/login", url.Values{
		"username": {username},
		"password": {password},
	})
	return http_session, r, http_err
}

func register_and_login(username string, password string) (*http.Client, *http.Response, error) {
	// Registers and logs in in one go
	register(username, password, "", "")
	return login(username, password)
}

func logout(http_session *http.Client) (*http.Response, error) {
	// Helper function to logout
	return http_session.Get(BASE_URL + "/logout") // Follows redirects by default
}

func add_message(http_session *http.Client, text string, t *testing.T) (*http.Response, error) {
	// Records a message
	r, err := http_session.PostForm(BASE_URL+"/add_message", url.Values{"text": {text}})
	if text != "" {
		defer r.Body.Close()
		r_text, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(r_text[:]), "Your message was recorded") {
			t.Fatalf("got \"" + string(r_text))
		}
	}
	return r, err
}

func TestMiniTwitTestCase(t *testing.T) {

}
