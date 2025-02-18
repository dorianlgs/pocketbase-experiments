package main

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pocketbase/pocketbase/tools/store"
)

type InMem struct {
	// TODO: it would be nice to have a mutex here
	// TODO: use pointers to avoid copying
	users    *store.Store[string, PasskeyUser]
	sessions *store.Store[string, webauthn.SessionData]

	log Logger
}

func (i *InMem) GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil

}

func NewInMem(log Logger) *InMem {

	return &InMem{
		users:    store.New[string, PasskeyUser](nil),
		sessions: store.New[string, webauthn.SessionData](nil),
		log:      log,
	}
}

func (i *InMem) GetSession(token string) (webauthn.SessionData, bool) {
	i.log.Printf("[DEBUG] GetSession: %v", i.sessions.Get(token))
	val, ok := i.sessions.GetOk(token)

	return val, ok
}

func (i *InMem) SaveSession(token string, data webauthn.SessionData) {
	i.log.Printf("[DEBUG] SaveSession: %s - %v", token, data)
	i.sessions.Set(token, data)
}

func (i *InMem) DeleteSession(token string) {
	i.log.Printf("[DEBUG] DeleteSession: %v", token)
	i.sessions.Remove(token)
}

func (i *InMem) GetOrCreateUser(userName string) PasskeyUser {
	i.log.Printf("[DEBUG] GetOrCreateUser: %v", userName)
	if _, ok := i.users.GetOk(userName); !ok {
		i.log.Printf("[DEBUG] GetOrCreateUser: creating new user: %v", userName)
		i.users.Set(userName, &User{
			ID:          []byte(userName),
			DisplayName: userName,
			Name:        userName,
		})
	}

	return i.users.Get(userName)
}

func (i *InMem) SaveUser(user PasskeyUser) {
	i.log.Printf("[DEBUG] SaveUser: %v", user.WebAuthnName())
	i.log.Printf("[DEBUG] SaveUser: %v", user)
	i.users.Set(user.WebAuthnName(), user)
}
