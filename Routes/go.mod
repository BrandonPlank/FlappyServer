module brandonplank.org/FlappyServer/routes

go 1.17

require (
	brandonplank.org/FlappyServer/database v0.0.0
	brandonplank.org/FlappyServer/global v0.0.0
	brandonplank.org/FlappyServer/models v0.0.0
	github.com/gofiber/fiber/v2 v2.29.0
	github.com/google/uuid v1.3.0
	github.com/jinzhu/gorm v1.9.16
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.34.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
)

replace (
	brandonplank.org/FlappyServer/database => ./../Database
	brandonplank.org/FlappyServer/global => ./../Global
	brandonplank.org/FlappyServer/models => ./../Models
)
