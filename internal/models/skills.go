package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Skill struct {
	Id   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
	Yrs  float64            `bson:"years"`
}

func (s Skill) ViewAttr() string {
	return fmt.Sprintf("id: %v, name: %s, yrs: %v", s.Id, s.Name, s.Yrs)
}

func NewSkill(name string, yrs float64) *Skill {
	return &Skill{
		Name: name,
		Yrs:  yrs,
	}
}
