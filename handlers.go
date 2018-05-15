package csync

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

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

	for k, v := range r.Form {
		err = insert(nativeCookie.Value, k, v[0])
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
