package global

import (
	"html/template"
	"io"
	"sync"
)

var (
	SecretToken   string
	OwnerOverride string
	Mutex         = &sync.Mutex{}
	Writer        io.Writer
	ViewDir       = "Resources/View/"
	BootStrap     *template.Template
)
