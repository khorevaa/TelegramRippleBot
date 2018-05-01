package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"io"
	"io/ioutil"
	"bytes"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"time"
	"github.com/ChimeraCoder/anaconda"
	"net/url"
)

var (
	bot           *tgbotapi.BotAPI
	configuration Config
	db            *gorm.DB
	phrases       map[int]string
	cache         CachedStats
	sinceTwitter  = make(map[string]int64)
	twitter       *anaconda.TwitterApi
)

func main() {
	initLog()
	initConfig()
	initStrings()
	initDB()
	initTwitter()
	initCache()

	var err error
	bot, err = tgbotapi.NewBotAPI(configuration.BotToken)
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	go checkTransactions()
	go checkTwitter()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			command := update.Message.Command()
			switch command {
			case "start":
				start(update.Message)
			case "addwallet":
				addWallet(update.Message)
			case "removewallet":
				removeWallet(update.Message)
			case "xrp":
				xrp(update.Message)
			}
		}
	}

}

func initLog() {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
}

func initConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	log.Print("First 10 bytes from config.json")
	log.Print(body[:10])
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	log.Print("First 10 bytes after trim")
	reader := bytes.NewReader(body)
	log.Print(body[:10])
	decoder := json.NewDecoder(reader)

	err = decoder.Decode(&configuration)
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}

}

func initStrings() {
	file, err := os.Open("strings.json")
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	log.Print("First 10 bytes from strings.json")
	log.Print(body[:10])
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	log.Print("First 10 bytes after trim")
	reader := bytes.NewReader(body)
	log.Print(body[:10])
	decoder := json.NewDecoder(reader)

	err = decoder.Decode(&phrases)
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}
}

func initDB() {
	var err error
	db, err = gorm.Open("sqlite3", "data.db")
	if err != nil {
		log.Print("********** ERROR: ")
		log.Panic("Failed to connect database")
	} else {
		log.Print("Opened DB")
	}
	db.LogMode(true)
	log.Print("Set LogMode")
	db.AutoMigrate(&User{},&Wallet{})
	log.Print("Migrated")
}

func initTwitter(){
	twitter = anaconda.NewTwitterApiWithCredentials(configuration.TwitterAccessToken,
		configuration.TwitterAccessSecret,
		configuration.TwitterConsumerKey,
		configuration.TwitterConsumerSecret)

	v := url.Values{}
	v.Add("exclude_replies", "true")
	v.Add("include_rts", "false")
	for _, val := range configuration.TwitterAccounts{
		v.Add("screen_name", val)
		v.Add("count", "1")
		searchResult, err := twitter.GetUserTimeline(v)
		if err != nil{
			log.Print(err)
		}
		if len(searchResult) > 0{
			sinceTwitter[val] = searchResult[0].Id
		}
	}
}

func initCache(){
	loadChart()
	cache = CachedStats{Time: time.Now(), Stats:getRippleStats()}
}