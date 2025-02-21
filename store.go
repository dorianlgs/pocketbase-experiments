package main

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/store"
)

type InMem struct {
	sessions *store.Store[string, LocalSession]
	log      Logger
	app      *pocketbase.PocketBase
}

func (i *InMem) GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil

}

func NewInMem(log Logger, app *pocketbase.PocketBase) *InMem {
	return &InMem{
		sessions: store.New[string, LocalSession](nil),
		log:      log,
		app:      app,
	}
}

func (i *InMem) GetSession(token string) (LocalSession, bool) {
	i.log.Printf("[DEBUG] GetSession: %v", i.sessions.Get(token))
	val, ok := i.sessions.GetOk(token)

	return val, ok
}

func (i *InMem) SaveSession(token string, data LocalSession) {
	i.log.Printf("[DEBUG] SaveSession: %s - %v", token, data)
	i.sessions.Set(token, data)
}

func (i *InMem) DeleteSession(token string) {
	i.log.Printf("[DEBUG] DeleteSession: %v", token)
	i.sessions.Remove(token)
}

func (i *InMem) GetOrCreateUser(email string) (PasskeyUser, error) {
	i.log.Printf("[DEBUG] GetOrCreateUser: %v", email)

	_, err := i.app.FindFirstRecordByData("users", "email", email)
	if err != nil {
		collection, err := i.app.FindCollectionByNameOrId("users")
		if err != nil {
			return nil, err
		}

		record := core.NewRecord(collection)

		record.Set("email", email)
		record.Set("name", email)
		record.SetPassword("Lorem ipsum")

		err = i.app.Save(record)
		if err != nil {
			return nil, err
		}
	}

	userRecord, userErr := i.app.FindFirstRecordByData("users", "email", email)
	if userErr != nil {
		return nil, userErr
	}

	user := &User{
		ID:          []byte(email),
		DisplayName: userRecord.GetString("name"),
		Name:        userRecord.GetString("name"),
		app:         i.app,
	}

	return user, nil

}

func (i *InMem) SaveUser(user PasskeyUser) {
	//i.log.Printf("[DEBUG] SaveUser: %v", user.WebAuthnName())
	//i.log.Printf("[DEBUG] SaveUser: %v", user)
	//i.users.Set(user.WebAuthnName(), user)
}
