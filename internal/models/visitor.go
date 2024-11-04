package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Visitor struct {
	Id             primitive.ObjectID `bson:"_id:omitempty"`
	VisitCount     int                `bson:"visit_count"`
	Address        string             `bson:"address"`
	HasCreatedTask bool               `bson:"has_created_task"`
}

func NewVisitor(visitCount int, addr string, hasCreatedTask bool) *Visitor {
	return &Visitor{
		VisitCount:     visitCount,
		Address:        addr,
		HasCreatedTask: hasCreatedTask,
	}
}

func (v Visitor) ViewAttr() string {
	return fmt.Sprintf("%#v\n", v)
}
