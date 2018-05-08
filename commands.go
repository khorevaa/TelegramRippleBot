package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
	cmc "github.com/coincircle/go-coinmarketcap"
	"log"
	"fmt"
	"time"
)

func start(message *tgbotapi.Message) {
	addUserIfAbsent(message.From)
	sendMessage(message.Chat.ID, phrases[0], nil)
}

func addWallet(message *tgbotapi.Message) {
	fields := strings.Fields(message.Text)
	if len(fields) >= 2 {
		if !checkAddress(fields[1]) {
			sendMessage(message.Chat.ID, phrases[4], nil)
			return
		}
		addWalletDB(message)
		var text string
		if len(fields) == 3 {
			text = fmt.Sprintf(phrases[1], fields[2])
		} else {
			text = fmt.Sprintf(phrases[1], fields[1])
		}
		sendMessage(message.Chat.ID, text, nil)
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
	for _, coin := range coins {
		newMap[coin.Rank] = coin
	}

	var currText string
	for i := 0; i < 10; i++ {
		number := i
		number++
		numberStr := numberEmojis[number]
		currText = numberStr + " " + newMap[number].Symbol + " " + float64ToString(newMap[number].PriceUSD) +
			" USD _(" + float64WithSign(newMap[number].PercentChange24H) + "%)_\n"

		if newMap[number].Name == "Ripple" {
			currText = "*" + currText + "*"
			currText = strings.Replace(currText, "_", "", -1)
		}
		text += currText
	}
	sendMessage(message.Chat.ID, text, nil)
}

func price(message *tgbotapi.Message) {
	fields := strings.Fields(message.Text)
	if len(fields) == 2{
		priceCoin(message)
	}else{
		priceXrp(message)
	}


}

func priceCoin(message *tgbotapi.Message){
	var text string
	fields := strings.Fields(message.Text)
	ticker := getCurrency(fields[1])
	if ticker != ""{
		coin, err := cmc.GetCoinData(ticker)
		if err != nil {
			log.Print(err)
		}

		text = "*"+coin.Symbol+" = " + float64ToString(coin.PriceUSD) + " USD* " +
			float64WithSign(coin.PercentChange24H) + "% _(24h)_" + "\n\n" +
			float64WithSign(coin.PercentChange7D) + "% _(7d)_"

	}else{
		text = phrases[5]
	}

	sendMessage(message.Chat.ID, text, nil)
}

func priceXrp(message *tgbotapi.Message){
	coin, err := cmc.GetCoinData("ripple")
	if err != nil {
		log.Print(err)
	}

	text := "*XRP = " + float64ToString(coin.PriceUSD) + " USD* " +
		float64WithSign(coin.PercentChange24H) + "% _(24h)_" + "\n\n" +
		float64WithSign(coin.PercentChange7D) + "% _(7d)_" +
		"\n\n📈 view /chart or more XRP /stats \n"+"👉 [Buy/Sell XRP]("+configuration.BuySellXRP+")"
	sendMessage(message.Chat.ID, text, nil)
}

func chart(message *tgbotapi.Message) {

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
