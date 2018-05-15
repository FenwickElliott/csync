package csync

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"time"

	"gopkg.in/mgo.v2/bson"
)

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

func check(err error) {
	if err != nil {
		panic(err)
	}
}
