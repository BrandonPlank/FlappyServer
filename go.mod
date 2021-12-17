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
