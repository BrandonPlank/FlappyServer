package global

import (
	"html/template"
	"io"
	"sync"
)

var (
	SECRET_TOKEN   string
	OWNER_OVERRIDE string
	Mutex          = &sync.Mutex{}
	Writer         io.Writer
	ViewDir        = "Resources/View/"
	BootStrap      *template.Template
)
