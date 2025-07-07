package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

type App struct {
	Router      *mux.Router
	Middlewares *Middleware
	config      *Env
}

type shortenReq struct {
	URL                 string `json:"url" validate:"required"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

type shortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

func (a *App) Initizlize(e *Env) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.config = e
	a.Router = mux.NewRouter()
	a.Middlewares = &Middleware{}
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// a.Router.HandleFunc("/api/shorten", a.createShorten).Methods("POST")
	// a.Router.HandleFunc("/api/info", a.getShortlinkInfo).Methods("GET")
	// a.Router.HandleFunc("/{shortlink:[a-z-Z0-9]{1,11}}", a.redirect).Methods("GET")
	m := alice.New(a.Middlewares.LogMiddleware, a.Middlewares.RecoverMiddleware)
	a.Router.Handle("/api/shorten", m.ThenFunc(a.createShorten)).Methods("POST")
	a.Router.Handle("/api/info", m.ThenFunc(a.getShortlinkInfo)).Methods("GET")
	a.Router.Handle("/{shortlink:[a-z-Z0-9]{1,11}}", m.ThenFunc(a.redirect)).Methods("GET")
}

func (a *App) createShorten(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest, fmt.Errorf("parse parmeters failed %v", r.Body)})
		return
	}

	if msg := ValidateParam(req); msg != "" {
		respondWithError(w, StatusError{http.StatusBadRequest, fmt.Errorf("parse parmeters failed %v", req)})
		return
	}
	defer r.Body.Close()

	s, err := a.config.S.Shorten(req.URL, req.ExpirationInMinutes)
	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJson(w, http.StatusCreated, shortlinkResp{Shortlink: s})
	}
}

func (a *App) getShortlinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	s := vals.Get("shortlink")

	d, err := a.config.S.ShortlinkInfo(s)
	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJson(w, http.StatusOK, d)
	}
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	u, err := a.config.S.Unshorten(vars["shortlink"])
	if err != nil {
		respondWithError(w, err)
	} else {
		http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	}
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
