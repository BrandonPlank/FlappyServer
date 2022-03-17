package routes

import (
	"brandonplank.org/FlappyServer/database"
	"brandonplank.org/FlappyServer/global"
	"brandonplank.org/FlappyServer/models"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/gofiber/fiber/v2"
	guuid "github.com/google/uuid"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

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
	var users []models.User
	database.DB.Where("is_banned=?", false).Find(&users)
	users = sortUsers(users)
	deaths := countDeaths(users)
	players := len(users)
	if len(users) > 25 {
		users = users[:25]
	}
	return ctx.Render("main", fiber.Map{
		"Users":   users,
		"players": players,
		"deaths":  deaths,
		"year":    time.Now().Format("2006"),
	})
}

func GetUser(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	var user models.User
	database.DB.Where("name=? ", name).First(&user)

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
		"year":      time.Now().Format("2006"),
	})
}

func Bans(ctx *fiber.Ctx) error {
	var users []models.User
	database.DB.Where("is_banned=?", true).Find(&users)

	return ctx.Render("bans", fiber.Map{
		"Users": users,
		"year":  time.Now().Format("2006"),
	})
}

// MARK: v1

func V1GetUsers(ctx *fiber.Ctx) error {
	var users []models.User
	database.DB.Where("is_banned=?", false).Find(&users)
	return ctx.JSON(models.ConvertUsersToPublicUsers(users))
}

func V1InternalUsers(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)

	if readUser.Admin || readUser.Owner {
		var users []models.User
		database.DB.Find(&users)
		return ctx.JSON(users)
	} else {
		return ctx.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}
}

func V1RegisterUser(ctx *fiber.Ctx) error {
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

	database.DB.Where("name=?", data["name"]).First(&readUser)

	if readUser.Name == data["name"] {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "User exists"})
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{
		Name:         data["name"],
		PasswordHash: string(password),
	}

	database.DB.Create(&user)

	return ctx.JSON(user)
}

func V1Login(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")

	var readUser models.User
	database.DB.First(&readUser, "name=?", name)

	// We need user accounts, so allow login to create accounts
	if readUser.ID == guuid.Nil || readUser.Name != name {
		// Create account
		err := V1RegisterUser(ctx)
		HandleError(err)

		// Do another query to get user
		database.DB.Where("name=?", name).First(&readUser)
		if readUser.ID == guuid.Nil || readUser.Name != name {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Undefined error"})
		}
	}
	return ctx.JSON(readUser)
}

func V1SubmitScore(ctx *fiber.Ctx) error {
	var data models.Score
	err := ctx.BodyParser(&data)
	HandleError(err)

	name := ctx.Locals("name")

	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)

	current := sha256.Sum256([]byte(strconv.Itoa(data.Score) + global.SecretToken + strconv.Itoa(data.Time)))
	currentVerify := hex.EncodeToString(current[:])

	if data.Verify != currentVerify {
		log.Println("[SCORE] Unable to verify score for", name)
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Unable to verify score"})
	}

	log.Println("[SCORE] Score verification passed for", name)
	log.Println("[SCORE] User:", name, "[ID:"+readUser.ID.String()+"] submitted score:", data.Score, "took", data.Time, "seconds.")

	if data.Time+100 < data.Score || data.Time-100 > data.Score && !readUser.Admin {
		database.DB.Model(&readUser).Update("is_banned", true)
		database.DB.Model(&readUser).Update("ban_reason", "Cheating (Anti cheat)")
		ip := global.GetIPFromContext(ctx)
		log.Println(fmt.Sprintf("[BAN] Banned %s with the IP of %s", name, ip))
		_, _ = global.BansClient.CreateEmbeds([]discord.Embed{
			{
				Title:       "Banned User",
				Description: fmt.Sprintf("%s has been banned for cheating by the ani-cheat"),
				Color:       16734296,
				Fields: []discord.EmbedField{
					{
						Name:  "Score submitted",
						Value: fmt.Sprintf("%d", data.Score),
					},
					{
						Name:  "Time",
						Value: fmt.Sprintf("%d seconds", data.Time),
					},
					{
						Name:  "IP",
						Value: ip,
					},
				},
				Footer: &discord.EmbedFooter{
					Text:    "Flappybird API",
					IconURL: "https://flappybird.brandonplank.org/images/favicon.png",
				},
			},
		})
	}

	if data.Score > readUser.Score {
		database.DB.Model(&readUser).Update("score", data.Score)
		log.Println("[SCORE] Processed score for", name)
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func V1SubmitDeath(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	database.DB.Model(&readUser).Update("deaths", readUser.Deaths+1)
	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func V1IsJailbroken(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	database.DB.Model(&readUser).Update("jailbroken", true)
	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func V1Emulator(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	database.DB.Model(&readUser).Update("ran_in_emulator", true)
	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func V1HasHackedTools(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	database.DB.Model(&readUser).Update("has_hacked_tools", true)
	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Success"})
}

func V1GetID(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	if readUser.ID == guuid.Nil || readUser.Name != name {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}
	return ctx.SendString(readUser.ID.String())
}

func V1UserCount(ctx *fiber.Ctx) error {
	var readUsers []models.User
	database.DB.Where("is_banned=?", false).Find(&readUsers)
	if len(readUsers) < 1 {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "No users in the database"})
	}
	return ctx.SendString(strconv.Itoa(len(readUsers)))
}

func V1GlobalDeaths(ctx *fiber.Ctx) error {
	var readUsers []models.User
	database.DB.Where("is_banned=?", false).Find(&readUsers)
	return ctx.SendString(strconv.Itoa(countDeaths(readUsers)))
}

func V1Leaderboard(ctx *fiber.Ctx) error {
	amountStr := ctx.Params("amount")
	amount, _ := strconv.Atoi(amountStr)
	var readUsers []models.User
	database.DB.Where("is_banned=?", false).Find(&readUsers)
	if len(readUsers) < amount {
		return ctx.JSON(models.ConvertUsersToPublicUsers(sortUsers(readUsers)))
	}
	return ctx.JSON(models.ConvertUsersToPublicUsers(sortUsers(readUsers[:amount])))
}

func V1RestoreScore(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	id := ctx.Params("id")
	scoreString := ctx.Params("score")
	score, _ := strconv.Atoi(scoreString)

	// Owner global override
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	if !readUser.Admin && !readUser.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	var user models.User
	database.DB.First(&user, "id=?", guuid.MustParse(id))
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

	log.Println("[RESTORE]", readUser.Name, "is restoring", user.Name+"'s score to", strconv.Itoa(score)+",was", user.Score)

	database.DB.Model(&user).Update("score", score)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Updated " + user.Name + "'s score"})
}

func V1Ban(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	reason, err := url.QueryUnescape(ctx.Params("reason"))
	if err != nil {
		return ctx.SendString(err.Error())
	}
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	if !readUser.Admin && !readUser.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	//OVERRIDE:
	var user models.User
	database.DB.First(&user, "id=?", guuid.MustParse(id))
	if user.ID == guuid.Nil || user.ID.String() != id {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}

	if user.Admin || user.Owner {
		if !readUser.Owner {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You cannot ban another admin"})
		}
	}

	format, _ := url.QueryUnescape(reason)

	database.DB.Model(&user).Update("is_banned", true)
	database.DB.Model(&user).Update("ban_reason", format)

	log.Println(fmt.Sprintf("[BAN] %s Banned %s: %s", readUser.Name, user.Name, format))
	_, _ = global.BansClient.CreateEmbeds([]discord.Embed{
		{
			Title:       fmt.Sprintf("%s banned %s", readUser.Name, user.Name),
			Description: fmt.Sprintf("%s has been banned for %s", user.Name, format),
			Color:       16734296,
			Footer: &discord.EmbedFooter{
				Text:    "Flappybird API",
				IconURL: "https://flappybird.brandonplank.org/images/favicon.png",
			},
		},
	})

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Banned " + user.Name})
}

func V1UnBan(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	if !readUser.Admin && !readUser.Owner {
		//if id == owner_override {
		//	goto OVERRIDE
		//}
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	//OVERRIDE:
	var user models.User
	database.DB.First(&user, "id=?", guuid.MustParse(id))
	if user.ID == guuid.Nil || user.ID.String() != id {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}

	if user.Admin || user.Owner {
		if !readUser.Owner {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You cannot unban another admin"})
		}
	}

	database.DB.Model(&user).Update("is_banned", false)
	database.DB.Model(&user).Update("ban_reason", "")

	log.Println(fmt.Sprintf("[UNBAN] %s unbanned %s", readUser.Name, user.Name))
	_, _ = global.BansClient.CreateEmbeds([]discord.Embed{
		{
			Title:       fmt.Sprintf("%s unbanned %s", readUser.Name, user.Name),
			Description: fmt.Sprintf("%s has been unbanned", user.Name),
			Color:       16734296,
			Footer: &discord.EmbedFooter{
				Text:    "Flappybird API",
				IconURL: "https://flappybird.brandonplank.org/images/favicon.png",
			},
		},
	})

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Unbanned " + user.Name})
}

func V1DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	if !readUser.Admin && !readUser.Owner {
		//if id == owner_override {
		//	goto OVERRIDE
		//}
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	//OVERRIDE:
	var user models.User
	database.DB.First(&user, "id=?", guuid.MustParse(id))
	if user.ID == guuid.Nil || user.ID.String() != id {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}

	if user.Admin || user.Owner {
		if !readUser.Owner {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You cannot delete another admin"})
		}
	}

	log.Println("[DELETE]", readUser.Name, "is deleting", user.Name)

	database.DB.Delete(&user).Where("id=?", guuid.MustParse(id))

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Deleted " + user.Name})
}

func V1MakeAdmin(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	if !readUser.Owner {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}
	id := ctx.Params("id")
	var user models.User
	database.DB.First(&user, "id=?", guuid.MustParse(id))
	if user.ID == guuid.Nil || user.ID.String() != id {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Failed to get user ID"})
	}

	database.DB.Model(&user).Update("admin", true)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Made " + user.Name + " an admin"})
}

func V1ServerLogFile(ctx *fiber.Ctx) error {
	name := ctx.Locals("name")
	var readUser models.User
	database.DB.Where("name=?", name).First(&readUser)
	if !readUser.Owner && !readUser.Admin {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	file, err := os.Open("flappyserver.log")
	HandleError(err)
	defer file.Close()
	var reader io.Reader
	reader = file
	return ctx.SendStream(reader)
}
