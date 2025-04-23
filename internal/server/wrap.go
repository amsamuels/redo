// internal/handler/wrap.go
package server

import (
	"log"
	"net/http"
)

func Wrap(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		h(w, r)
	}
}
