package main

import (
	// "fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/TheAimHero/sb/internal/assert"
)

func TestPing(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	code, _, body := ts.get(t, "/ping")
	assert.Equal[string](t, string(body), "OK")
	assert.Equal[int](t, code, http.StatusOK)
}

func TestSnippetView(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{name: "Valid ID", urlPath: "/snippet/view/1", wantCode: http.StatusOK, wantBody: "An old silent pond..."},
		{name: "Non-existent ID", urlPath: "/snippet/view/2", wantCode: http.StatusNotFound},
		{name: "Negative ID", urlPath: "/snippet/view/-1", wantCode: http.StatusNotFound},
		{name: "Decimal ID", urlPath: "/snippet/view/1.23", wantCode: http.StatusNotFound},
		{name: "String ID", urlPath: "/snippet/view/foo", wantCode: http.StatusNotFound},
		{name: "Empty ID", urlPath: "/snippet/view/", wantCode: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			assert.Equal[int](t, code, tt.wantCode)
			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestUserSignUp(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	_, _, body := ts.get(t, "/user/signup")
	validCSRFToken := extractCSRFToken(t, body)
	const (
		validName     = "Bob"
		validPassword = "validPa$$word"
		validEmail    = "bob@example.com"
		formTag       = "<form action='/user/signup' method='POST' novalidate>"
	)
	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Invalid CSRF Token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "Empty name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty email",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid email",
			userName:     validName,
			userEmail:    "bob@example.",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Short password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "pa$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Duplicate email",
			userName:     validName,
			userEmail:    "dupe@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)
			code, _, body := ts.postForm(t, "/user/signup", form)
			assert.Equal[int](t, code, tt.wantCode)
			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}
}

func TestSnippetCreate(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	t.Run("Unauthenticated", func(t *testing.T) {
		code, _, _ := ts.get(t, "/snippet/create")
		assert.Equal[int](t, code, http.StatusSeeOther)
	})
	t.Run("Authenticated", func(t *testing.T) {
		code, _, body := ts.get(t, "/user/login")
		assert.Equal[int](t, code, http.StatusOK)
		csrfToken := extractCSRFToken(t, body)
		form := url.Values{
			"email":      {"alice@example.com"},
			"password":   {"pa$$word"},
			"csrf_token": {csrfToken},
		}
		code, _, _ = ts.postForm(t, "/user/login", form)
		assert.Equal[int](t, code, http.StatusSeeOther)
		code, _, body = ts.get(t, "/snippet/create")
		assert.Equal[int](t, code, http.StatusOK)
		assert.StringContains(t, body, "<form action='/snippet/create' method='POST'>")
	})
}