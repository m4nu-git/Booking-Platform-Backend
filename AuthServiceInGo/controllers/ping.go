package controllers

import "net/http"

func PingHandlers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong"))
}