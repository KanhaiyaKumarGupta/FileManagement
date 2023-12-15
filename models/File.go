package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileTransaction struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	FileName   string             `bson:"fileName"`
	Operation  string             `bson:"operation"`
	Size       int64              `bson:"size,omitempty"`
	AccessedAt time.Time          `bson:"accessedAt"`
}
