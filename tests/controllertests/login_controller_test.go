package controllertests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/go-playground/assert.v1"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignIn(t *testing.T) {
	err := refreshUserTable()
	if err != nil {
		log.Fatal(err)
	}

	user, err := seedOneUser()

	if err != nil {
		fmt.Printf("this is the error %v\n", err)
	}

	samples := []struct {
		email        string
		password     string
		errorMessage string
	}{
		{
			email:        user.Email,
			password:     "password", //Note the password has to be this, not the hashed one from the database
			errorMessage: "",
		},
		{
			email:        user.Email,
			password:     "Wrong password",
			errorMessage: "crypto/bcrypt: hashedPassword is not the hash of the given password",
		},
		{
			email:        "Wrong email",
			password:     "password",
			errorMessage: "record not found",
		},
	}

	for _, v := range samples {
		token, err := server.SignIn(v.email, v.password)

		if err != nil {
			assert.Equal(t, err, errors.New(v.errorMessage))
		} else {
			assert.NotEqual(t, token, "")
		}
	}
}

func TestLogin(t *testing.T) {
	err := refreshUserTable()
	if err != nil {
		return
	}

	_, err = seedOneUser()

	if err != nil {
		fmt.Printf("this is the error %v\n", err)
	}

	samples := []struct {
		inputJSON    string
		statusCode   int
		email        string
		password     string
		errorMessage string
	}{
		{
			inputJSON:    `{"email": "pet@gmail.com", "password": "password"}`,
			statusCode:   200,
			errorMessage: "",
		},
		{
			inputJSON:    `{"email": "pet@gmail.com", "password": "wrong password"}`,
			statusCode:   422,
			errorMessage: "Incorrect password",
		},
		{
			inputJSON:    `{"email": "frank@gmail.com", "password": "password"}`,
			statusCode:   422,
			errorMessage: "Incorrect details",
		},
		{
			inputJSON:    `{"email": "kangmail.com", "password": "password"}`,
			statusCode:   422,
			errorMessage: "Invalid email provided",
		},
		{
			inputJSON:    `{"email": "", "password": "password"}`,
			statusCode:   422,
			errorMessage: "Required email",
		},
		{
			inputJSON:    `{"email": "kan@gmail.com", "password": ""}`,
			statusCode:   422,
			errorMessage: "Required password",
		},
		{
			inputJSON:    `{"email": "", "password": "password"}`,
			statusCode:   422,
			errorMessage: "Required email",
		},
	}

	for _, v := range samples {

		req, err := http.NewRequest("POST", "/login", bytes.NewBufferString(v.inputJSON))
		if err != nil {
			t.Errorf("this is the error: %v", err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.Login)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, v.statusCode)

		if v.statusCode == 200 {
			assert.NotEqual(t, rr.Body.String(), "")
		}

		if v.statusCode == 422 && v.errorMessage != "" {
			responseMap := make(map[string]interface{})
			err = json.Unmarshal([]byte(rr.Body.String()), &responseMap)

			if err != nil {
				t.Errorf("cannot convert to json: %v", err)
			}

			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}
