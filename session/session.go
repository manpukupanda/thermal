package session

import (
	"io"
	"thermal/model"
)

type Session struct {
	Manifest *model.Manifest
	Instance *model.XBRLInstance
	Schema   *model.XBRLSchema
	Stdout   io.Writer
	Stderr   io.Writer
}
