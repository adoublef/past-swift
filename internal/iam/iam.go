package iam

import (
	"github.com/adoublef/past-swift/internal/oauth2"
	"github.com/rs/xid"
)

type User struct {
	Profile  *Profile  `json:"profile"`
	OAuth2ID oauth2.ID `json:"oauth2"`
}

func (u User) ID() xid.ID { return u.Profile.ID }

func NewUser(oauthID oauth2.ID, login, photo, name string) *User {
	u := User{
		Profile: &Profile{
			ID:    xid.New(),
			Login: login,
			Photo: photo,
			Name:  name,
		},
		OAuth2ID: oauthID,
	}

	return &u
}

type Profile struct {
	ID xid.ID `json:"id"`
	// Login is the email of the user
	Login string `json:"login"`
	// Photo is the avatar associated with the user
	Photo string `json:"photoUrl"`
	// Name is the user's full name
	Name string `json:"name"`
}
