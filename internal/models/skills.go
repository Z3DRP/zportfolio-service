package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Skill struct {
	Id   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
	Yrs  float32            `bson:"years"`
}

func (s Skill) ViewAttr() string {
	return fmt.Sprintf("id: %v, name: %s, yrs: %s", s.Id, s.Name, s.Yrs)
}

func NewSkill(name string, yrs float32) *Skill {
	return &Skill{
		Name: name,
		Yrs:  yrs,
	}
}
