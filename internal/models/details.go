package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Detail struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Title       string             `bson:"title"`
	ExpOverview string             `bson:"experience_overview"`
	Quote       string             `bson:"quote"`
	Email       string             `bson:"email"`
	Phone       string             `bson:"phone"`
}

func (d Detail) ViewAttr() string {
	return fmt.Sprintf("id: %v, name: %s, title: %s, overview: %s", d.Id, d.Name, d.Title, d.ExpOverview)
}

func NewDetail(id, name, title, overview, quote, email, phone string) *Detail {
	return &Detail{
		Name:        name,
		Title:       title,
		ExpOverview: overview,
		Quote:       quote,
		Email:       email,
		Phone:       phone,
	}
}
