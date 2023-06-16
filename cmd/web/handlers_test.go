package main

import (
	"final-project/data"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var pageTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
	handler            http.HandlerFunc
	sessionData        map[string]any
	expectedHTML       string
}{
	{
		name:               "Home",
		url:                "/",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
		handler:            testApp.HomePage,
	},
	{
		name:               "Login Page",
		url:                "/login",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
		handler:            testApp.LoginPage,
		expectedHTML:       `<h1 class="mt-5">Login</h1>`,
	},
	{
		name:               "Logout",
		url:                "/logout",
		method:             "GET",
		expectedStatusCode: http.StatusSeeOther,
		handler:            testApp.Logout,
		sessionData: map[string]any{
			"userID": 1,
			"user":   data.User{},
		},
	},
}

func Test_Pages(t *testing.T) {
	pathToTemplates = "./templates"

	for _, entry := range pageTests {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest(entry.method, entry.url, nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		if len(entry.sessionData) > 0 {
			for key, value := range entry.sessionData {
				testApp.Session.Put(ctx, key, value)
			}
		}

		// response request
		// handler := http.HandlerFunc(testApp.HomePage)
		entry.handler.ServeHTTP(rr, req)

		if rr.Code != entry.expectedStatusCode {
			t.Errorf("%s Failed! Expected %d but got %d", entry.name, entry.expectedStatusCode, rr.Code)
		}

		if len(entry.expectedHTML) > 0 {
			html := rr.Body.String()
			if !strings.Contains(html, entry.expectedHTML) {
				t.Errorf("%s Failed! Expected HTML %s not found! ", entry.name, entry.expectedHTML)
			}
		}
	}
}

func TestConfig_PostLoginPage(t *testing.T) {
	pathToTemplates = "./templates"

	postedData := url.Values{
		"email":    {"admin@example.com"},
		"password": {"abc123abc123abc123"},
	}

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	handler := http.HandlerFunc(testApp.PostLoginPage)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Error("Wrong code returned!")
	}

	if !testApp.Session.Exists(ctx, "userID") {
		t.Error("Did not find userID in the session")
	}
}

func TestConfig_SubscribeToPlan(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/subscribe?id=1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.Session.Put(ctx, "user", data.User{
		ID:        1,
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		Active:    1,
	})

	handler := http.HandlerFunc(testApp.SubscribeToPlan)
	handler.ServeHTTP(rr, req)

	testApp.Wait.Wait()

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected StatusSeeOther but got %d", rr.Code)
	}
}
