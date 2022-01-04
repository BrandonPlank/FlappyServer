package models

import (
	"brandonplank.org/FlappyServer/database"
	"golang.org/x/crypto/bcrypt"
	"time"

	//"github.com/jinzhu/gorm"
	//"github.com/gofiber/fiber/v2"
	guuid "github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type User struct {
	ID        guuid.UUID `gorm:"primary_key" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name             string `json:"name" gorm:"unique"`
	Score            int    `json:"score"`
	Deaths           int    `json:"deaths"`
	PasswordHash     string `json:"passwordHash"`
	Jailbroken       bool   `json:"jailbroken"`
	HasHackedTools   bool   `json:"hasHackedTools"`
	RanInEmulator    bool   `json:"ranInEmulator"`
	HasModifiedScore bool   `json:"hasModifiedScore"`
	IsBanned         bool   `json:"isBanned"`
	BanReason        string `json:"banReason"`
	Admin            bool   `json:"admin"`
	Owner            bool   `json:"owner"`
}

type PublicUser struct {
	ID        guuid.UUID `gorm:"primary_key" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name      string `json:"name"`
	Score     int    `json:"score"`
	Deaths    int    `json:"deaths"`
	IsBanned  bool   `json:"isBanned"`
	BanReason string `json:"banReason"`
}

func (base *User) BeforeCreate(tx *gorm.DB) (err error) {
	uuid := guuid.New()
	base.ID = uuid
	return
}

func ConvertUserToPublicUser(user *User) PublicUser {
	return PublicUser{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Name: user.Name, IsBanned: user.IsBanned, BanReason: user.BanReason, Score: user.Score, Deaths: user.Deaths}
}

func ConvertUsersToPublicUsers(users []User) (publicUsers []PublicUser) {
	for i := 0; i < len(users); i++ {
		publicUsers = append(publicUsers, ConvertUserToPublicUser(&users[i]))
	}
	return
}

type AuthErr struct {
	Success bool
	Message string
}

func Auth(name string, password string) bool {

	var user User

	database.DB.Where("name=?", name).First(&user)

	if user.ID == guuid.Nil {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))

	if err != nil {
		return false
	}
	return true
}
