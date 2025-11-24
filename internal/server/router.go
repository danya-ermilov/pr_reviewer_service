package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type RouterHolder struct {
    Router *mux.Router
}

func (r *RouterHolder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.Router.ServeHTTP(w, req)
}

func (r *RouterHolder) Mux() *mux.Router {
    return r.Router
}

func NewRouter() *RouterHolder {
    return &RouterHolder{Router: mux.NewRouter()}
}
