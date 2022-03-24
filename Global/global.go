package global

import (
	"github.com/DisgoOrg/disgo/webhook"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"html/template"
	"io"
	"strings"
	"sync"
)

var (
	BansClient    *webhook.Client
	APIClient     *webhook.Client
	SecretToken   string
	OwnerOverride string
	Mutex         = &sync.Mutex{}
	Writer        io.Writer
	ViewDir       = "Resources/View/"
	BootStrap     *template.Template
)

func GetIPFromContext(ctx *fiber.Ctx) string {
	ip := string(ctx.Request().Header.Peek("X-Forwarded-For"))
	if len(ip) < 1 {
		ip = string(ctx.Request().Header.Peek("CF-Connecting-IP"))
		if len(ip) < 1 {
			ip = ctx.IP()
			if len(ip) < 4 {
				ip = ctx.Hostname()
				if len(ip) < 4 {
					ip = "None"
				}
			}
		}
	}
	if strings.Contains(ip, ",") {
		ip = strings.Split(ip, ", ")[0]
	}
	return ip
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
