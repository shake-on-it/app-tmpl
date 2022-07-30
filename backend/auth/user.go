package auth

import (
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"hash"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	UserTypeGuest  = ""
	UserTypeMe     = "me"
	UserTypeAdmin  = "admin"
	UserTypeNormal = "normal"

	UserStatusUnverified = ""
	UserStatusVerified   = "verified"
	UserStatusPrivileged = "privileged"
)

type User struct {
	ID       primitive.ObjectID   `bson:"_id" json:"id"`
	Name     string               `bson:"name" json:"name"`
	Email    string               `bson:"email" json:"email"`
	Type     string               `bson:"type" json:"type,omitempty"`
	Status   string               `bson:"status" json:"status,omitempty"`
	Sessions []primitive.ObjectID `bson:"sessions" json:"-"`
}

func (u *User) Validate() error {
	if u.ID == primitive.NilObjectID {
		u.ID = primitive.NewObjectID()
	}
	if u.Name == "" {
		return errors.New("must have name")
	}
	if u.Email == "" {
		return errors.New("must have email")
	}
	if u.Sessions == nil {
		u.Sessions = []primitive.ObjectID{}
	}
	return nil
}

func (u *User) Valid() error {
	return nil
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c Credentials) Validate() error {
	if c.Username == "" {
		return errors.New("must have username")
	}
	if c.Password == "" {
		return errors.New("must have password")
	}
	return nil
}

type Registration struct {
	Credentials
	Email string `json:"email"`
}

const (
	DigestTypeEmpty  = ""
	DigestTypeSHA256 = "sha256"
	DigestTypeSHA512 = "sha512"
)

type Password struct {
	ID             primitive.ObjectID `bson:"_id"`
	Username       string             `bson:"username"`
	Salt           primitive.Binary   `bson:"salt"`
	HashedPassword primitive.Binary   `bson:"hashed_password"`
	Iterations     int                `bson:"iterations"`
	KeyLength      int                `bson:"key_length"`
	DigestType     string             `bson:"digest_type"`
}

func (p *Password) Validate() error {
	if p.ID == primitive.NilObjectID {
		p.ID = primitive.NewObjectID()
	}
	if p.Username == "" {
		return errors.New("must have username")
	}
	if len(p.Salt.Data) == 0 {
		return errors.New("must have salt")
	}
	if len(p.HashedPassword.Data) == 0 {
		return errors.New("must be hashed")
	}
	if p.Iterations == 0 {
		return errors.New("must have iterations")
	}
	if p.KeyLength == 0 {
		return errors.New("must have key length")
	}
	if p.DigestType == DigestTypeEmpty {
		return errors.New("must have digest type")
	}
	return nil
}

func DigestHash(digestType string) func() hash.Hash {
	if digestType == DigestTypeSHA512 {
		return sha512.New
	}
	return sha256.New
}
