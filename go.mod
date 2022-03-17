module brandonplank.org/FlappyServer

go 1.17

require (
	brandonplank.org/FlappyServer/database v0.0.0
	brandonplank.org/FlappyServer/global v0.0.0
	brandonplank.org/FlappyServer/models v0.0.0
	brandonplank.org/FlappyServer/routes v0.0.0
)

replace (
	brandonplank.org/FlappyServer/database => ./Database
	brandonplank.org/FlappyServer/global => ./Global
	brandonplank.org/FlappyServer/models => ./Models
	brandonplank.org/FlappyServer/routes => ./Routes
)

require (
	github.com/DisgoOrg/disgo v0.7.2
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/gofiber/fiber/v2 v2.29.0
	github.com/gofiber/template v1.6.20
	github.com/jinzhu/gorm v1.9.16
	github.com/joho/godotenv v1.4.0
)

require (
	github.com/DisgoOrg/log v1.1.3 // indirect
	github.com/DisgoOrg/snowflake v1.0.4 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sasha-s/go-csync v0.0.0-20210812194225-61421b77c44b // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.34.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
)
