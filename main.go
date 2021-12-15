package main

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/common-nighthawk/go-figure"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html"
	guuid "github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"kayladev.me/FlappyServer/database"
	"kayladev.me/FlappyServer/models"
)

const port = 8069

var (
	secret_token   string
	owner_override string
	mutex          = &sync.Mutex{}
	writer         io.Writer
	ViewDir        = "Resources/View/"
	bootstrap      *template.Template
)

func initDatabase() {
	var err error
	database.DatabaseConnection, err = gorm.Open("sqlite3", "flappybird.db")
	HandleError(err)
	log.Println("[DATABASE] Connection Opened to Database")
	database.DatabaseConnection.AutoMigrate(&models.User{})
}

func setupRoutes(app *fiber.App) {
	app.Use(
		logger.New(
			logger.Config{
				Format:     "${time} [${method}]->${status} Latency->${latency} - ${path} | ${error}\n",
				TimeFormat: "2006/01/02 15:04:05",
				Output:     writer,
			},
		),
		cors.New(cors.Config{
			AllowCredentials: true,
		}),
		func(ctx *fiber.Ctx) error {
			ctx.Append("Access-Control-Allow-Origin", "*")
			ctx.Append("Developer", "crypticplank")
			ctx.Append("License", "BSD 3-Clause License")
			ctx.Append("Source-Url", "https://github.com/crypticplank/FlappyServer")
			return ctx.Next()
		},
	)

	app.Static("/", "./Public")

	v1 := app.Group("/v1")

	authv1 := v1.Group("/auth")
	authv1.Use(basicauth.New(basicauth.Config{
		Authorizer:      models.Auth,
		ContextUsername: "name",
	}))

	app.Get("/", Home)
	app.Get("/bans", Bans)
	app.Get("/user/:name", GetUser)

	v1.Get("/leaderboard/:amount", Leaderboard)
	v1.Get("/globalDeaths", GlobalDeaths)
	v1.Get("/userCount", UserCount)
	v1.Get("/users", GetUsers)
	v1.Get("/getID/:name", GetID)
	v1.Post("/registerUser", RegisterUser)

	authv1.Post("/submitScore", SubmitScore)
	authv1.Post("/submitDeath", SubmitDeath)
	authv1.Post("/isJailbroken", IsJailbroken)
	authv1.Post("/emulator", Emulator)
	authv1.Post("/hasHackedTools", HasHackedTools)
	authv1.Post("/login", Login)

	authv1.Get("/internalUsers", InternalUsers)
	authv1.Get("/ban/:id/:reason", Ban)
	authv1.Get("/makeAdmin/:id", MakeAdmin)
	authv1.Get("/unban/:id", UnBan)
	authv1.Get("/delete/:id", DeleteUser)
	authv1.Get("/restoreScore/:id/:score", RestoreScore)
}

func main() {
	myFigure := figure.NewFigure("FlappyBird Server", "", true)
	myFigure.Print()

	log.Println("[START] Starting the FlappyBird REST server")

	// Setup dotenv
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[ERROR] Error loading .env file")
	}

	secret_token = os.Getenv("SECRET_TOKEN")

	owner_override = os.Getenv("GLOBAL_OWNER_OVERRIDE_KEY")
	log.Println("[START] Got secret token:", secret_token)
	log.Println("[START] Got owner override key: " + owner_override[:4] + "***************************")

	// Setup Logging
	file, err := os.OpenFile("flappyserver.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	HandleError(err)
	defer file.Close()
	writer = io.MultiWriter(os.Stdout, file)
	log.SetOutput(writer)

	//Setup views
	engine := html.New("./Resources/Views", ".html")
	//engine.Reload(true)
	//engine.Debug(true)

	router := fiber.New(fiber.Config{DisableStartupMessage: true, Views: engine})

	initDatabase()

	log.Println("[START] Setting up routes")
	setupRoutes(router)

	log.Println("[START] Starting server on port", strconv.Itoa(port))
	log.Fatalln(router.Listen(":" + strconv.Itoa(port)))

	defer database.DatabaseConnection.Close()
}

func countDeaths(users []models.User) int {
	total := 0
	for i := 0; i < len(users); i++ {
		total += users[i].Deaths
	}
	return total
}

func sortUsers(a []models.User) []models.User {
	for i := 0; i < len(a)-1; i++ {
		for j := i + 1; j < len(a); j++ {
			if a[i].Score <= a[j].Score {
				temp := a[i]
				a[i] = a[j]
				a[j] = temp
			}
		}
	}
	return a
}

/*
 *	Routes
 */

func HandleError(err error) bool {
	if err != nil {
		log.Println("[ERROR]", err)
		return true
	}
	return false
}

func Home(ctx *fiber.Ctx) error {
	db := database.DatabaseConnection
	var users []models.User
	db.Where("is_banned=?", false).Find(&users)
	users = users[:24]
	return ctx.Render("main", fiber.Map{
		"Users":   sortUsers(users),
		"players": len(users),
		"deaths":  countDeaths(users),
	})
}

func GetUser(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	var user models.User
	database.DatabaseConnection.Where("name=? ", name).First(&user)

	gotUser := true

	if user.Name != name {
		gotUser = false
	}

	return ctx.Render("user", fiber.Map{
		"name":      user.Name,
		"isBanned":  user.IsBanned,
		"banReason": user.BanReason,
		"score":     user.Score,
		"deaths":    user.Deaths,
		"id":        user.ID.String(),
		"user":      gotUser,
	})
}

func Bans(ctx *fiber.Ctx) error {
	db := database.DatabaseConnection
	var users []models.User
	db.Where("is_banned=?", true).Find(&users)

	return ctx.Render("bans", fiber.Map{
		"Users": users,
	})
}

func GetUsers(ctx *fiber.Ctx) error {
	db := database.DatabaseConnection
	var users []models.User
	db.Where("is_banned=?", false).Find(&users)
	return ctx.JSON(models.ConvertUsersToPublicUsers(users))
}

func InternalUsers(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)

	if readUser.Admin || readUser.Owner {
		db := database.DatabaseConnection
		var users []models.User
		db.Find(&users)
		return ctx.JSON(users)
	} else {
		return ctx.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}
}

func RegisterUser(ctx *fiber.Ctx) error {
	var data map[string]string
	err := ctx.BodyParser(&data)
	HandleError(err)

	if len(data["name"]) < 1 || len(data["password"]) < 1 {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "User did not provide username or password"})
	}

	if len(data["name"]) > 15 {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Username is to long"})
	}

	if strings.Contains(data["name"], " ") {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Name cannot have any whitespaces"})
	}

	log.Println("[REGISTER]", data["name"], "is registering")

	var readUser models.User

	database.DatabaseConnection.Where("name=?", data["name"]).First(&readUser)

	if readUser.Name == data["name"] {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "User exists"})
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{
		Name:         data["name"],
		PasswordHash: string(password),
	}

	database.DatabaseConnection.Create(&user)

	return ctx.JSON(user)
}

func Login(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")

	var readUser models.User
	database.DatabaseConnection.First(&readUser, "name=?", name)

	// We need user accounts, so allow login to create accounts
	if readUser.ID == guuid.Nil || readUser.Name != name {
		// Create account
		err := RegisterUser(ctx)
		HandleError(err)

		// Do another query to get user
		database.DatabaseConnection.Where("name=?", name).First(&readUser)
		if readUser.ID == guuid.Nil || readUser.Name != name {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Undefined error"})
		}
	}
	return ctx.JSON(readUser)
}

func SubmitScore(ctx *fiber.Ctx) error {
	var data models.Score
	err := ctx.BodyParser(&data)
	HandleError(err)

	name := ctx.Locals("name")

	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)

	current := sha256.Sum256([]byte(strconv.Itoa(data.Score) + secret_token + strconv.Itoa(data.Time)))
	currentVerify := hex.EncodeToString(current[:])

	if data.Verify != currentVerify {
		log.Println("[SCORE] Unable to verify score for", name)
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Unable to verify score"})
	}

	log.Println("[SCORE] Score verification passed for", name)
	log.Println("[SCORE] User:", name, "[ID:"+readUser.ID.String()+"] submitted score:", data.Score, "took", data.Time, "seconds.")

	if data.Time+100 < data.Score || data.Time-100 > data.Score && !readUser.Admin {
		database.DatabaseConnection.Model(&readUser).Update("is_banned", true)
		database.DatabaseConnection.Model(&readUser).Update("ban_reason", "Cheating (Anti cheat)")
	}

	if data.Score > readUser.Score {
		database.DatabaseConnection.Model(&readUser).Update("score", data.Score)
		log.Println("[SCORE] Processed score for", name)
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func SubmitDeath(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")

	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)

	database.DatabaseConnection.Model(&readUser).Update("deaths", readUser.Deaths+1)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func IsJailbroken(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")

	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)

	database.DatabaseConnection.Model(&readUser).Update("is_jailbroken", true)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func Emulator(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)
	database.DatabaseConnection.Model(&readUser).Update("ran_in_emulator", true)
	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func HasHackedTools(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)
	database.DatabaseConnection.Model(&readUser).Update("has_hacked_tools", true)
	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func GetID(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)
	if readUser.ID == guuid.Nil || readUser.Name != name {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}
	return ctx.SendString(readUser.ID.String())
}

func UserCount(ctx *fiber.Ctx) error {
	var readUsers []models.User
	database.DatabaseConnection.Where("is_banned=?", false).Find(&readUsers)
	if len(readUsers) < 1 {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "No users in the database"})
	}
	return ctx.SendString(strconv.Itoa(len(readUsers)))
}

func GlobalDeaths(ctx *fiber.Ctx) error {
	var readUsers []models.User
	database.DatabaseConnection.Where("is_banned=?", false).Find(&readUsers)
	return ctx.SendString(strconv.Itoa(countDeaths(readUsers)))
}

func Leaderboard(ctx *fiber.Ctx) error {
	amountStr := ctx.Params("amount")
	amount, _ := strconv.Atoi(amountStr)
	var readUsers []models.User
	database.DatabaseConnection.Where("is_banned=?", false).Find(&readUsers)
	if len(readUsers) < amount {
		return ctx.JSON(sortUsers(readUsers))
	}
	return ctx.JSON(models.ConvertUsersToPublicUsers(sortUsers(readUsers[:amount])))
}

func RestoreScore(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	id := ctx.Params("id")
	scoreString := ctx.Params("score")
	score, _ := strconv.Atoi(scoreString)

	// Owner global override
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)
	if !readUser.Admin && !readUser.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	var user models.User
	database.DatabaseConnection.First(&user, id)
	if user.ID == guuid.Nil || user.ID.String() != id {
		// Owner global override
		//if id == owner_override {
		//	goto OVERRIDE
		//}
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}
	//OVERRIDE:
	if !readUser.Owner && user.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "That's illegal, this incident will be recorded"})
	}

	database.DatabaseConnection.Model(&user).Update("score", score)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Updated " + user.Name + "'s score"})
}

func Ban(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	reason := ctx.Params("reason")
	name := ctx.Locals("name")
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)
	if !readUser.Admin || !readUser.Owner {
		// Owner global override
		//if id == owner_override {
		//	goto OVERRIDE
		//}
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	//OVERRIDE:
	var user models.User
	database.DatabaseConnection.First(&user, id)
	if user.ID == guuid.Nil || user.ID.String() != id {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}

	if user.Owner || user.Admin {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You cannot remove another admin"})
	}

	database.DatabaseConnection.Model(&user).Update("is_banned", true)
	database.DatabaseConnection.Model(&user).Update("ban_reason", reason)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Made " + user.Name + " an admin"})
}

func UnBan(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	name := ctx.Locals("name")
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)
	if !readUser.Admin && !readUser.Owner {
		//if id == owner_override {
		//	goto OVERRIDE
		//}
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	//OVERRIDE:
	var user models.User
	database.DatabaseConnection.First(&user, id)
	if user.ID == guuid.Nil || user.ID.String() != id {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}

	if user.Owner || user.Admin {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You cannot remove another admin"})
	}

	database.DatabaseConnection.Model(&user).Update("is_banned", false)
	database.DatabaseConnection.Model(&user).Update("ban_reason", "")

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Made " + user.Name + " an admin"})
}

func DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	name := ctx.Locals("name")
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)
	if !readUser.Admin && !readUser.Owner {
		//if id == owner_override {
		//	goto OVERRIDE
		//}
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	//OVERRIDE:
	var user models.User
	database.DatabaseConnection.First(&user, id)
	if user.ID == guuid.Nil || user.ID.String() != id {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}
	if !user.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You cannot remove another admin"})
	}
	database.DatabaseConnection.Delete(&user).Where("id=?", id)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Deleted " + user.Name})
}

func MakeAdmin(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DatabaseConnection.Where("name=?", name).First(&readUser)
	if !readUser.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	id := ctx.Params("id")
	var user models.User
	database.DatabaseConnection.First(&user, id)
	if user.ID == guuid.Nil || user.ID.String() != id {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}

	database.DatabaseConnection.Model(&user).Update("admin", true)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Made " + user.Name + " an admin"})
}
