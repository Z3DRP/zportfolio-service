package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id      primitive.ObjectID `bson:"_id:omitempty"`
	Uid     string             `bson:"uid"`
	Name    string             `bson:"name"`
	Company string             `bson:"company"`
	Email   string             `bson:"email"`
	Phone   string             `bson:"phone"`
	Roles   []string           `bson:"roles"`
}

func NewUser(uid, nm, cm, em, ph string, rols []string) *User {
	return &User{
		Uid:     uid,
		Name:    nm,
		Company: cm,
		Email:   em,
		Phone:   ph,
		Roles:   rols,
	}
}

func (u User) ViewAttr() string {
	return fmt.Sprintf("%#v\n", u)
}
