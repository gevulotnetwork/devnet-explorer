package api

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Store interface {
	Hello() string
}

type API struct {
	r *httprouter.Router
	s Store
}

func New(s Store) *API {
	a := &API{
		r: httprouter.New(),
		s: s,
	}
	a.bind()
	return a
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.ServeHTTP(w, r)
}

func (a *API) bind() {
	a.r.GET("/", a.helloWorld)
}

func (a *API) helloWorld(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, a.s.Hello())
}
