package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/TheAimHero/sb/internal/models"
	"github.com/TheAimHero/sb/internal/validator"
	"github.com/julienschmidt/httprouter"
)

type snippetCreateFrom struct {
	Title   string
	Content string
	Expires int
	validator.Validator
}

type userSignupForm struct {
	Name     string
	Email    string
	Password string
	validator.Validator
}

type userLoginForm struct {
	Email    string
	Password string
	validator.Validator
}

type PasswordUpdateFrom struct {
	currentPassword         string
	newPassword             string
	newPasswordConfirmation string
	validator.Validator
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.Snippets = snippets
	app.render(w, http.StatusOK, "home.html", data)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "about.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if id == 0 || err != nil {
		app.notFound(w)
		return
	}
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	data := app.newTemplateData(r)
	data.Snippet = snippet
	app.render(w, http.StatusOK, "view.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.Form.Get("title")
	content := r.Form.Get("content")
	expires, err := strconv.Atoi(r.Form.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := snippetCreateFrom{
		Title:   title,
		Content: content,
		Expires: expires,
	}
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field is too long (maximum is 100 characters)")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.html", data)
		return
	}
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = snippetCreateFrom{
		Expires: 365,
	}
	app.render(w, http.StatusOK, "create.html", data)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{
		Name:     "",
		Email:    "",
		Password: "",
	}
	app.render(w, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	name := r.Form.Get("name")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	form := userSignupForm{
		Name:     name,
		Email:    email,
		Password: password,
	}
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{
		Email:    "",
		Password: "",
	}
	app.render(w, http.StatusOK, "login.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	form := userLoginForm{
		Email:    email,
		Password: password,
	}
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
	return
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	userId := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(userId)
	data.User = user
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.render(w, http.StatusOK, "account.html", data)
}

func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	form := &PasswordUpdateFrom{
		currentPassword:         "",
		newPassword:             "",
		newPasswordConfirmation: "",
	}
	data := app.newTemplateData(r)
	data.Form = form
	app.render(w, http.StatusOK, "password.html", data)
}

func (app *application) accountPasswordUpdatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}
	currentPassword := r.Form.Get("currentPassword")
	newPassword := r.Form.Get("newPassword")
	newPasswordConfirmation := r.Form.Get("newPasswordConfirmation")
	form := PasswordUpdateFrom{
		currentPassword:         currentPassword,
		newPassword:             newPassword,
		newPasswordConfirmation: newPasswordConfirmation,
	}
	form.CheckField(validator.NotBlank(form.currentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.newPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.newPasswordConfirmation), "newPasswordConfirmation", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.newPassword, 8), "newPassword", "This field must be at least 8 characters long")
	form.CheckField(validator.Equals[string](form.newPassword, form.newPasswordConfirmation), "newPasswordConfirmation", "Passwords do not match")
	if !form.Valid() {
		data := app.newTemplateData(r)
		form.currentPassword = ""
		form.newPassword = ""
		form.newPasswordConfirmation = ""
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "password.html", data)
		return
	}
	userId := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	err = app.users.PasswordUpdate(userId, form.currentPassword, form.newPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("currentPassword", "Current password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "password.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
