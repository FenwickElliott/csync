package csync

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/mgo.v2"
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
	http.HandleFunc("/out", out)
	http.HandleFunc("/forward", forward)
	http.HandleFunc("/back", back)
	fmt.Println("Serving:", service.Name, "on port:", service.Port)
	return http.ListenAndServe(":"+service.Port, nil)
}
