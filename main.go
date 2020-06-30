package main

import (
	"fmt"
	recaptcha "github.com/dpapathanasiou/go-recaptcha"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

var file *os.File

func submit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	recaptchaResponse := r.FormValue("g-recaptcha-response")
	ip := getRealIp(r)
	result, err := recaptcha.Confirm(ip, recaptchaResponse)
	if err != nil || result != true {
		if err != nil {
			log.Println("recaptcha server error", err)
		}

		c := &http.Cookie{Name: "form-submit", Value: "notok"}
		http.SetCookie(w, c)

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")

	if _, err := file.WriteString(fmt.Sprintf("%s,%s\n", email, username)); err != nil {
		log.Println(err)
	}

	c := &http.Cookie{Name: "form-submit", Value: "ok"}
	http.SetCookie(w, c)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func main() {
	captchaSecret := os.Getenv("RECAPTCHA_SECRET")
	listenAddr := os.Getenv("LISTEN_ADDR")

	r := mux.NewRouter()

	recaptcha.Init(captchaSecret)

	var err error
	file, err = os.OpenFile("users.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	r.HandleFunc("/submit", submit).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("src/")))

	srv := &http.Server{
		Handler:      handlers.RecoveryHandler()(handlers.CompressHandler(handlers.LoggingHandler(os.Stdout, r))),
		Addr:         listenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
