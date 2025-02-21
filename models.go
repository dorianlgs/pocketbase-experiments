package main

import (
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	b64 "encoding/base64"
)

type User struct {
	ID          []byte
	DisplayName string
	Name        string
	creds       []webauthn.Credential
	app         *pocketbase.PocketBase
}

func (o *User) WebAuthnID() []byte {
	return o.ID
}

func (o *User) WebAuthnName() string {
	return o.Name
}

func (o *User) WebAuthnDisplayName() string {
	return o.DisplayName
}

func (o *User) WebAuthnIcon() string {
	return "https://pics.com/avatar.png"
}

func (o *User) WebAuthnCredentials() []webauthn.Credential {

	userRecord, err := o.app.FindFirstRecordByData("users", "email", string(o.ID))

	if err != nil {
		return nil
	}

	userId := userRecord.GetString("id")

	records, err := o.app.FindAllRecords("credentials",
		dbx.NewExp("user_id = {:user_id}", dbx.Params{"user_id": userId}),
	)

	if err != nil {
		return nil
	}

	var credentials = []webauthn.Credential{}

	for _, record := range records {

		var result webauthn.Credential
		err = record.UnmarshalJSONField("json_credential", &result)
		if err != nil {
			return nil
		}
		credentials = append(credentials, result)
	}

	return credentials

}

func (o *User) AddCredential(credential *webauthn.Credential, email string) error {

	collection, err := o.app.FindCollectionByNameOrId("credentials")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)

	transports, err := json.Marshal(credential.Transport)
	if err != nil {
		return err
	}

	userRecord, err := o.app.FindFirstRecordByData("users", "email", email)
	if err != nil {
		return err
	}

	//	fmt.Printf("signature_count: %s", string(o.ID))

	json_credential, err := json.Marshal(credential)
	if err != nil {
		return err
	}

	credential_id := b64.StdEncoding.EncodeToString(credential.ID)
	public_key := b64.StdEncoding.EncodeToString(credential.PublicKey)
	aaguid := b64.StdEncoding.EncodeToString(credential.Authenticator.AAGUID)

	record.Set("user_id", string(userRecord.GetString("id")))
	record.Set("credential_id", credential_id)
	record.Set("public_key", public_key)
	record.Set("attestation_type", string(credential.AttestationType))
	record.Set("aaguid", aaguid)
	record.Set("signature_count", credential.Authenticator.SignCount)
	record.Set("last_used_date", time.Now())
	record.Set("type", credential.Descriptor().Type)
	record.Set("transports", string(transports))
	record.Set("backup_eligible", credential.Flags.BackupEligible)
	record.Set("backup_state", credential.Flags.BackupState)
	record.Set("json_credential", json_credential)

	err = o.app.Save(record)
	if err != nil {
		return err
	}

	return nil

}

func (o *User) UpdateCredential(credential *webauthn.Credential) error {

	record, err := o.app.FindRecordById("credentials", string(credential.ID))
	if err != nil {
		return err
	}

	transports, err := json.Marshal(credential.Transport)
	if err != nil {
		return err
	}

	json_credential, err := json.Marshal(credential)
	if err != nil {
		return err
	}

	public_key := b64.StdEncoding.EncodeToString(credential.PublicKey)
	aaguid := b64.StdEncoding.EncodeToString(credential.Authenticator.AAGUID)

	record.Set("public_key", public_key)
	record.Set("attestation_type", string(credential.AttestationType))
	record.Set("aaguid", aaguid)
	record.Set("signature_count", credential.Authenticator.SignCount)
	record.Set("last_used_date", time.Now())
	record.Set("type", credential.Descriptor().Type)
	record.Set("transports", string(transports))
	record.Set("backup_eligible", credential.Flags.BackupEligible)
	record.Set("backup_state", credential.Flags.BackupState)
	record.Set("json_credential", json_credential)

	err = o.app.Save(record)
	if err != nil {
		return err
	}

	return nil
}
