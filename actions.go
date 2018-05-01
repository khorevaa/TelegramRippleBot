package main

import (
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func sendMessage(chatId int64, text string, keyboard interface{}){
	msg := tgbotapi.NewMessage(chatId, text)

	_, ok := keyboard.(tgbotapi.ReplyKeyboardMarkup)
	if ok{
		msg.ReplyMarkup = keyboard
	}else{
		_, ok = keyboard.(tgbotapi.InlineKeyboardMarkup)
		if ok{
			msg.ReplyMarkup = &keyboard
		}else {
			msg.ReplyMarkup = nil
		}
	}

	_, err := bot.Send(msg)
	if err != nil {
		log.Print(err)
	}
	log.Printf("[Bot] SENT %s TO %v", msg.Text, msg.ChatID)
}

func sendPhoto(photo tgbotapi.PhotoConfig) string {
	response, err := bot.Send(photo)
	if err != nil {
		log.Print(err)
		return ""
	}
	log.Printf("[Bot] PHOTO %s TO %v", photo.FileID, photo.ChatID)
	return (*(response.Photo))[0].FileID
}

func setUploadingPhoto(id int64){
	_, err := bot.Send(tgbotapi.NewChatAction(id, tgbotapi.ChatUploadPhoto))
	if err != nil {
		log.Print(err)
	}
}
