package csync

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	db      *mgo.Database
	service Service
)

type association struct {
	NativeCookie  string
	PartnerCookie string
}

// Service defines the case specific parameters
type Service struct {
	Name        string
	Port        string
	MongoServer string
	Redirect    string
}

// Serve opens the service
func Serve(service Service) error {
	if service.Name == "" {
		return errors.New("A service name must be provided")
	}
	if service.Port == "" {
		service.Port = "80"
	}

	session, err := mgo.Dial(service.MongoServer)
	check(err)
	defer session.Close()
	db = session.DB(service.Name)

	http.HandleFunc("/in", in)
	fmt.Println("Serving on port:", service.Port)
	return http.ListenAndServe(":"+service.Port, nil)
}

func in(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	partnerID := r.FormValue("partner")
	partnerCookie := r.FormValue("cookieID")

	nativeCookie, err := r.Cookie(service.Name + "ID")
	if nativeCookie == nil {
		nativeCookie = setCookie(&w, r)
	} else {
		check(err)
	}

	res := association{}
	c := db.C(partnerID)
	err = c.Find(bson.M{service.Name + "id": nativeCookie.Value}).One(&res)
	if err != nil {
		c.Insert(association{nativeCookie.Value, partnerCookie})
		err = c.Find(bson.M{service.Name + "id": nativeCookie.Value}).One(&res)
	}
	check(err)
	if res.PartnerCookie != partnerCookie {
		panic("partnerCookie doesn't match")
	}

	// implement redirect
	http.Redirect(w, r, service.Redirect+"/in?partner="+service.Name+"&cookieID="+nativeCookie.Value, 307)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func setCookie(w *http.ResponseWriter, r *http.Request) *http.Cookie {
	h := sha1.New()
	h.Write([]byte(time.Now().String() + r.RemoteAddr))
	cookie := http.Cookie{Name: service.Name + "ID", Value: hex.EncodeToString(h.Sum(nil)), Expires: time.Now().Add(365 * 24 * time.Hour)}
	http.SetCookie(*w, &cookie)
	return &cookie
}
