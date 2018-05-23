package main

import (
	"net/http"
	"log"
	"io/ioutil"
	cmc "github.com/coincircle/go-coinmarketcap"
	"strings"
	"telegram-bot-api"
	"time"
)

func getRippleStats() string {
	coin, err := cmc.Ticker(&cmc.TickerOptions{
		Symbol:  "XRP",
		Convert: "USD",
	})
	if err != nil {
		log.Print(err)
	}
	return "*Price:* " + float64ToString(coin.Quotes["USD"].Price) + " USD"
}

func checkAddress(a string) bool {
	if string(a[0]) != "r" {
		return false
	}
	if len(a) > 35 || len(a) < 25 {
		return false
	}
	return true
}

func getCurrency(name string) string {
	for _, listing := range listings {
		if strings.ToLower(listing.Name) == strings.ToLower(name) ||
			strings.ToLower(listing.Symbol) == strings.ToLower(name) {
			return listing.Symbol
		}
	}
	return ""
}

func getBalance(address string) float64 {
	resp, err := http.Get(config.RippleUrlBase + address + "/balances")
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	xrp := json.Get(bodyBytes, "balances", 0, "value").ToString()
	return stringToFloat64(xrp)
}

func rememberPost(message *tgbotapi.Message) {
	currPost.Message = *message
	currState = "waitingForDelay"
	sendMessage(message.Chat.ID, phrases[10], nil)
}

func rememberDelay(message *tgbotapi.Message) {
	currPost.DelayHours = stringToFloat64(message.Text)
	currPost.PostTime = time.Now()
	currState = "waitingForDestination"
	sendMessage(message.Chat.ID, phrases[11], nil)
}


func rememberDestination(message *tgbotapi.Message) {
	var posts []PendingPost
	readJson(&posts, "posts.json")
	currPost.Destination = stringToInt64(message.Text)
	posts = append(posts, currPost)
	writeJson(&posts, "posts.json")
	currPost = PendingPost{}
	currState = ""
	sendMessage(message.Chat.ID, phrases[19], nil)
}