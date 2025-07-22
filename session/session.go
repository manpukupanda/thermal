package session

import (
	"thermal/model"
)

type Session struct {
	Manifest *model.Manifest
	Instance *model.XBRLInstance
	Schema   *model.XBRLSchema
}
