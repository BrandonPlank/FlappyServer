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
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/gofiber/fiber/v2 v2.23.0
	github.com/gofiber/template v1.6.20
	github.com/jinzhu/gorm v1.9.16
	github.com/joho/godotenv v1.4.0
)

require (
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.13.4 // indirect
	github.com/mattn/go-sqlite3 v1.14.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.31.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
)
