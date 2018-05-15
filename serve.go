package csync

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	c            *mgo.Collection
	err          error
	service      Service
	nativeCookie http.Cookie
	partners     map[string]partner
)

type partner struct {
	Name       string
	AuthHeader string
	Address    string
	Scope      []string
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
	partners = getPartners()
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
	c = session.DB(service.Name).C("master")

	http.HandleFunc("/in", in)
	http.HandleFunc("/forward", forward)
	http.HandleFunc("/back", back)
	fmt.Println("Serving:", service.Name, "on port:", service.Port)
	return http.ListenAndServe(":"+service.Port, nil)
}

func in(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	partner := r.FormValue("partner")
	partnerCookie := r.FormValue("cookie")

	nativeCookie, err := r.Cookie(service.Name + "ID")
	if nativeCookie == nil {
		var res bson.M
		err = c.Find(bson.M{partner: partnerCookie}).One(&res)
		if err == nil {
			nativeCookie = setCookie(&w, r, res["_id"].(string))
		} else {
			nativeCookie = setCookie(&w, r, "new")
		}
	} else {
		check(err)
	}

	err = insert(nativeCookie.Value, partner, partnerCookie)
	check(err)

	if service.Redirect != "" && service.Redirect != partner {
		var res bson.M
		c.FindId(nativeCookie.Value).One(&res)

		str := partners[service.Redirect].Address + "/forward?"
		for k, v := range res {
			str += k + "=" + v.(string) + "&"
		}
		str += "back=" + service.Name
		str = strings.Replace(str, "_id", service.Name, -1)
		http.Redirect(w, r, str, 307)
	}
}

func forward(w http.ResponseWriter, r *http.Request) {
	fmt.Println("called: forward")
	r.ParseForm()
	nativeCookie, err := r.Cookie(service.Name + "ID")
	if nativeCookie == nil {
		nativeCookie = setCookie(&w, r, "new")
	} else {
		check(err)
	}
	for _, c := range r.Cookies() {
		err = insert(nativeCookie.Value, c.Name, c.Value)
		check(err)
	}

	str := partners[r.FormValue("back")].Address + "/back?partner=" + service.Name + "&cookie=" + nativeCookie.Value
	http.Redirect(w, r, str, 307)
}

func back(w http.ResponseWriter, r *http.Request) {
	fmt.Println("called back")
	r.ParseForm()
	partner := r.FormValue("partner")
	partnerCookie := r.FormValue("cookie")

	nativeCookie, err := r.Cookie(service.Name + "ID")
	if nativeCookie == nil {
		nativeCookie = setCookie(&w, r, "new")
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
	} else {
		return err
	}

	err = c.Find(bson.M{"_id": nativeID, partner: partnerCookie}).One(&res)
	return err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func setCookie(w *http.ResponseWriter, r *http.Request, cookieVal string) *http.Cookie {
	if cookieVal == "new" {
		h := sha1.New()
		h.Write([]byte(time.Now().String() + r.RemoteAddr))
		cookieVal = hex.EncodeToString(h.Sum(nil))
	}

	cookie := http.Cookie{Name: service.Name + "ID", Value: cookieVal, Expires: time.Now().Add(365 * 24 * time.Hour)}
	http.SetCookie(*w, &cookie)
	return &cookie
}
