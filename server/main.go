package main

import (
	"log"
	"net/http"

	"github.com/sunkink29/tictactoe/server/game"
)

type appHandler func(http.ResponseWriter, *http.Request) error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func init() {
	http.Handle("/host", appHandler(game.Host))
	http.Handle("/new", appHandler(game.New))
	http.Handle("/join", appHandler(game.Join))
	http.Handle("/checkHost", appHandler(game.CheckForJoin))
	fs := http.FileServer(http.Dir("../client"))
	http.Handle("/", fs)
}

func main() {
	log.Fatal(http.ListenAndServe(":8080", nil))
}
