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
	db           *mgo.Database
	c            *mgo.Collection
	err          error
	service      Service
	nativeCookie http.Cookie
)

type association struct {
	NativeCookie  string
	Partner       string
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
func Serve(serviceVars Service) error {
	service = serviceVars
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
	c = session.DB(service.Name).C("master")

	http.HandleFunc("/in", in)
	fmt.Println("Serving on port:", service.Port)
	return http.ListenAndServe(":"+service.Port, nil)
}

func in(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	partner := r.FormValue("partner")
	partnerCookie := r.FormValue("cookie")
	// fmt.Println(partner, partnerCookie)

	nativeCookie, err := r.Cookie(service.Name + "ID")
	if nativeCookie == nil {
		nativeCookie = setCookie(&w, r)
	} else {
		check(err)
	}
	err = insert(nativeCookie.Value, partner, partnerCookie)
	check(err)
}

func insert(nativeID, partner, partnerCookie string) error {
	var res bson.M

	err = c.Find(bson.M{"_id": nativeID, partner: partnerCookie}).One(&res)
	if err == nil {
		return err
	}

	err = c.FindId(nativeID).One(&res)
	if err == nil {
		err = c.UpdateId(nativeID, bson.M{"$set": bson.M{partner: partnerCookie}})
	} else if err.Error() == "not found" {
		err = c.Insert(bson.M{"_id": nativeID, partner: partnerCookie})
	}
	return err
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
