package session

import (
	"io"
	"thermal/model"
)

type Session struct {
	Manifest *model.Manifest
	Instance *model.XBRLInstance
	Schema   *model.XBRLSchema
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
}
