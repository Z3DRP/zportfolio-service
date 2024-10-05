package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Experience struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Company     string             `bson:"company"`
	Description string             `bson:"description"`
	Length      float64            `bson:"length"`
}

func (e Experience) ViewAttr() string {
	return fmt.Sprintf("id: %v, company: %s, description: %s, length: %v", e.Id, e.Company, e.Description, e.Length)
}

func NewExperience(comp, descrp string, length float64) *Experience {
	return &Experience{
		Company:     comp,
		Description: descrp,
		Length:      length,
	}
}
