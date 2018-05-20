package main

import (
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func addUserIfAbsent(user *tgbotapi.User) {
	log.Print("Started adding user")
	var u User
	db.Find(&u, "id = ?", user.ID)
	if u.ID == (User{}.ID) {
		userInsert := User{ID: int64(user.ID),
			FirstName: user.FirstName,
			LastName: user.LastName,
			UserName: user.UserName,
			Currency: "USD"}
		db.Create(&userInsert)
		log.Printf("Added new user with id = %v", userInsert.ID)
	}
}

func addWalletDB(message *tgbotapi.Message) {
	log.Print("Started adding wallet")
	var addr, name string
	fields := strings.Fields(message.Text)
	addr = fields[1]
	name = fields[2]


	u := getUser(message.From.ID)
	var wallet Wallet
	db.Find(&wallet, "address = ?", addr)
	if wallet.ID == 0 {
		db.Model(&u).Association("Wallets").Append(Wallet{Address: addr})
		db.First(&wallet, "address = ?", addr)
	} else {
		db.Model(&u).Association("Wallets").Append(wallet)
	}

	db.Model(&UserWallet{}).Where("user_id = ? AND wallet_id = ?",
			u.ID, wallet.ID).Update("name", name)

}

func resetWalletsDB(message *tgbotapi.Message) {
	log.Print("Started removing wallet")
	u := getUser(message.From.ID)
	var wallets []Wallet
	db.Model(&u).Association("Wallets").Find(&wallets)
	db.Model(&u).Association("Wallets").Clear()
	for _, wallet := range wallets {
		if db.Model(&wallet).Association("Users").Count() == 0 {
			db.Delete(&wallet)
		}
	}
}

func getUser(id int) User {
	log.Print("Started get user")
	var u User
	db.Find(&u, "id = ?", id)
	return u
}

func sendAllUsers(msg tgbotapi.MessageConfig){
	rows, _ := db.Table("users").Rows()
	for rows.Next() {
		var user User
		db.ScanRows(rows, &user)
		msg.ChatID = user.ID
		bot.Send(msg)
	}
}