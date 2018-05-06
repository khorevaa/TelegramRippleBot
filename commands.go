package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
	"strings"
	cmc "github.com/coincircle/go-coinmarketcap"
	"log"
)

func start(message *tgbotapi.Message) {
	addUserIfAbsent(message.From)
	sendMessage(message.Chat.ID, phrases[0], nil)
}

func addWallet(message *tgbotapi.Message) {
	if len(strings.Fields(message.Text)) >= 2 {
		addWalletDB(message)
		sendMessage(message.Chat.ID, phrases[1], nil)
	} else {
		sendMessage(message.Chat.ID, phrases[3], nil)
	}

}

func resetWallets(message *tgbotapi.Message) {
	resetWalletsDB(message)
	sendMessage(message.Chat.ID, phrases[2], nil)
}

func index(message *tgbotapi.Message) {
	var text string
	coins, err := cmc.GetAllCoinData(10)
	if err != nil {
		log.Print(err)
	}
	newMap := make(map[int]cmc.Coin)
	for _, coin := range coins{
		newMap[coin.Rank] = coin
	}

	var currText string
	for i := 0; i < 10; i++{
		number := i
		number++
		numberStr := numberEmojis[number]
		if newMap[number].PercentChange24H >= 0{
			currText = numberStr + " " + newMap[number].Symbol + " " + float64ToString(newMap[number].PriceUSD) +
				" USD _(+" + float64ToString(newMap[number].PercentChange24H) + "%)_\n"
		}else {
			currText = numberStr + " " + newMap[number].Symbol + " " + float64ToString(newMap[number].PriceUSD) +
				" USD _(" + float64ToString(newMap[number].PercentChange24H) + "%)_\n"
		}
		if newMap[number].Name == "Ripple"{
			currText = "*" + currText + "*"
			currText = strings.Replace(currText, "_", "", -1)
		}
		text += currText
	}
	sendMessage(message.Chat.ID, text, nil)
}

func xrp(message *tgbotapi.Message) {
	var photo tgbotapi.PhotoConfig
	setUploadingPhoto(message.Chat.ID)
	if time.Now().Sub(cache.Time).Minutes() > 3 || cache.PhotoId == "" {
		loadChart()
		cache.Time = time.Now()
		cache.Stats = getRippleStats()
		photo = tgbotapi.NewPhotoUpload(message.Chat.ID, "chart-usdt-xrp.png")
		photo.Caption = cache.Stats
		if id := sendPhoto(photo); id != "" {
			cache.PhotoId = id
		}
	} else {
		photo = tgbotapi.NewPhotoShare(message.Chat.ID, cache.PhotoId)
		sendPhoto(photo)
	}

}
