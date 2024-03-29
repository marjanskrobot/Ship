package routes

import (
	"fmt"
	"net/http"

	"github.com/learnToCrypto/lakoposlati/internal/logger"
	"github.com/learnToCrypto/lakoposlati/internal/platform/postgres"
	"github.com/learnToCrypto/lakoposlati/internal/provider"
	"github.com/learnToCrypto/lakoposlati/internal/sessions"
	"github.com/learnToCrypto/lakoposlati/internal/user"
)

// GET /login
// Show the login page
func Login(writer http.ResponseWriter, request *http.Request) {
	generateHTML(writer, nil, "layout/login.base", "public/navbar", "login")
}

// GET /signup
// Show the signup page
func SignupChoice(writer http.ResponseWriter, request *http.Request) {
	generateHTML(writer, nil, "layout/login.base", "public/navbar", "signup/choice")
}

// GET /signup
// Show the signup page
func Signup(writer http.ResponseWriter, request *http.Request) {
	if request.PostFormValue("user-type") == "customer" {
		generateHTML(writer, nil, "layout/login.base", "public/navbar", "signup/customer")
	} else if request.PostFormValue("user-type") == "provider" {
		generateHTML(writer, nil, "layout/login.base", "public/navbar", "signup/provider")
	}
}

// POST /signup
// Create the user account
func SignupAccount(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		logger.Danger(err, "Cannot parse form")
	}
	user := user.User{
		FirstName: request.PostFormValue("first_name"),
		LastName:  request.PostFormValue("last_name"),
		Username:  request.PostFormValue("username"),
		Email:     request.PostFormValue("email"),
		Password:  request.PostFormValue("password"),
		Role:      "user",
	}
	if err := user.Create(); err != nil {
		logger.Danger(err, "Cannot create user")
	}
	http.Redirect(writer, request, "/login", 302)
}

// TOdo how to save licence on server (not in database)
func SignupProvider(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		logger.Danger(err, "Cannot parse form")
	}

	userI := user.User{
		FirstName: request.PostFormValue("first_name"),
		LastName:  request.PostFormValue("last_name"),
		Username:  request.PostFormValue("username"),
		Email:     request.PostFormValue("email"),
		Password:  request.PostFormValue("password"),
		Role:      "provider",
	}

	if err := userI.Create(); err != nil {
		logger.Danger(err, "Cannot create user")
	}

	//get User ID
	userI, err = user.UserByEmail(userI.Email)
	//fmt.Println("user created", userI)

	providerI := provider.Provider{
		UserId:         userI.Id,
		MobilePhone:    request.PostFormValue("mobile_phone"),
		CompanyName:    request.PostFormValue("company_name"),
		CompanyAddr:    request.PostFormValue("company_addr"),
		CompanyCity:    request.PostFormValue("company_city"),
		CompanyZip:     request.PostFormValue("company_zip"),
		CompanyCountry: request.PostFormValue("company_country"),
	}

	fmt.Println("provider present,but not created", providerI)

	if err := providerI.Create(); err != nil {
		logger.Danger(err, "Cannot create provider")
	}
	http.Redirect(writer, request, "/login", 302)

}

/*
	file, header, err := request.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")
	fmt.Printf("File name %s\n", name[0])
	// Copy the file data to my buffer
	io.Copy(&Buf, file)
	// do something with the contents...
	// I normally have a struct defined and unmarshal into a struct, but this will
	// work as an example
	contents := Buf.String()
	fmt.Println(contents)
	// I reset the buffer in case I want to use it again
	// reduces memory allocations in more intense projects
	Buf.Reset()

*/

// POST /authenticate
// Authenticate the user given the email and password
func Authenticate(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	user, err := user.UserByEmail(request.PostFormValue("email"))
	if err != nil {
		http.Error(writer, "Cannot find user", http.StatusForbidden)
	}
	if user.Password == postgres.Encrypt(request.PostFormValue("password")) {
		sess, err := sessions.CreateSession(&user)
		if err != nil {
			logger.Danger(err, "Cannot create session")
		}
		cookie := http.Cookie{
			Name:     "_cookie",
			Value:    sess.Uuid,
			HttpOnly: true,
			Path:     "/",
			MaxAge:   3600 * 8, // in sec    3600 * 8   is 8 hours
		}
		http.SetCookie(writer, &cookie)
		//fmt.Println(user, "is logged in")
		http.Redirect(writer, request, "/", 302)
	} else {
		http.Error(writer, "Username and/or password do not match 2", http.StatusForbidden)
		http.Redirect(writer, request, "/login", 302)
	}

}

// GET /logout
// Logs the user out
func Logout(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("_cookie")
	fmt.Println("_cookie to delete:", cookie)
	if err != http.ErrNoCookie {
		//logger.Warning(err, "Failed to get cookie")
		session := sessions.Session{Uuid: cookie.Value}
		fmt.Println("session to delete in logout:", session)
		err := session.DeleteByUUID()
		fmt.Println("err in deleting session in logout:", err)
	}
	http.Redirect(writer, request, "/", 302)
}
