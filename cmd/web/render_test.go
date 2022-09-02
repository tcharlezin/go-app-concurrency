package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfig_AddDefaultData(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.Session.Put(ctx, "flash", "flash")
	testApp.Session.Put(ctx, "warning", "warning")
	testApp.Session.Put(ctx, "error", "error")

	td := testApp.addDefaultData(&TemplateData{}, req)

	if td.Flash != "flash" {
		t.Errorf("Failed to get flash data!")
	}

	if td.Warning != "warning" {
		t.Errorf("Failed to get warning data!")
	}

	if td.Error != "error" {
		t.Errorf("Failed to get error data!")
	}
}

func TestConfig_IsAuthenticated(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	auth := testApp.isAuthenticated(req)

	if auth {
		t.Error("Is not authenticated")
	}

	testApp.Session.Put(ctx, "userID", 1)

	auth = testApp.isAuthenticated(req)

	if !auth {
		t.Error("Is authenticated")
	}
}

func TestConfig_render(t *testing.T) {
	pathToTemplates = "./templates"

	rr := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.render(rr, req, "home.page.gohtml", &TemplateData{})

	if rr.Code != 200 {
		t.Error("Failed to render page")
	}
}
