package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

func start(message *tgbotapi.Message) {
	addUserIfAbsent(message.From)
	sendMessage(message.Chat.ID, phrases[0], nil)
}

func addWallet(message *tgbotapi.Message) {
	addWalletDB(message)
	sendMessage(message.Chat.ID, phrases[1], nil)
}

func removeWallet(message *tgbotapi.Message) {
	removeWalletDB(message)
	sendMessage(message.Chat.ID, phrases[2], nil)
}

func xrp(message *tgbotapi.Message){
	var photo tgbotapi.PhotoConfig
	setUploadingPhoto(message.Chat.ID)
	if time.Now().Sub(cache.Time).Minutes() > 3 || cache.PhotoId == ""{
		loadChart()
		cache.Time =  time.Now()
		cache.Stats = getRippleStats()
		photo = tgbotapi.NewPhotoUpload(message.Chat.ID, "chart-usdt-xrp.png")
		photo.Caption = cache.Stats
		if id := sendPhoto(photo); id != "" {
			cache.PhotoId = id
		}
	}else {
		photo = tgbotapi.NewPhotoShare(message.Chat.ID, cache.PhotoId)
		sendPhoto(photo)
	}

}