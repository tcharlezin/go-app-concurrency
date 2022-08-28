package main

import "net/http"

func (app *Config) SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}

func (app *Config) Auth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !app.Session.Exists(request.Context(), "userID") {
			app.Session.Put(request.Context(), "error", "Log in first!")
			http.Redirect(writer, request, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
