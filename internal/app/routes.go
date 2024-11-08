package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func initRoutes(router *chi.Mux) {
	router.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("alive"))
	})
}
