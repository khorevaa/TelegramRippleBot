package main

import (
	"telegram-bot-api"
	"log"
	"os"
	"io"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"time"
	"github.com/ChimeraCoder/anaconda"
	"net/url"
	"github.com/m90/go-chatbase"
	"github.com/fadion/gofixerio"
	cmc "github.com/coincircle/go-coinmarketcap"
	"github.com/coincircle/go-coinmarketcap/types"
)

var (
	bot                *tgbotapi.BotAPI
	config             Configuration
	db                 *gorm.DB
	metric             *chatbase.Client
	phrases            map[int]string
	cache24h           = CachedStats{PhotoIds: make(map[string]string)}
	cache30d           = CachedStats{PhotoIds: make(map[string]string)}
	sinceTwitter       = make(map[string]int64)
	twitter            *anaconda.TwitterApi
	listings           []*types.Listing
	currState          string
	currPost           PendingPost
	fixer              *fixerio.Request
	start1Keyboard     tgbotapi.InlineKeyboardMarkup
	start2Keyboard     tgbotapi.InlineKeyboardMarkup
	start3Keyboard     tgbotapi.InlineKeyboardMarkup
	txKeyboard         tgbotapi.InlineKeyboardMarkup
	priceKeyboard      tgbotapi.InlineKeyboardMarkup
	checkPriceKeyboard tgbotapi.InlineKeyboardMarkup
	statsKeyboard      tgbotapi.InlineKeyboardMarkup
	indexKeyboard      tgbotapi.InlineKeyboardMarkup
	balanceKeyboard    tgbotapi.InlineKeyboardMarkup
	chart30dKeyboard   tgbotapi.InlineKeyboardMarkup
	chart24hKeyboard   tgbotapi.InlineKeyboardMarkup
	yesNoKeyboard      tgbotapi.InlineKeyboardMarkup
	numberEmojis       = map[int]string{
		1:  "1⃣",
		2:  "2️⃣",
		3:  "3️⃣",
		4:  "4️⃣",
		5:  "5️⃣",
		6:  "6️⃣",
		7:  "7️⃣",
		8:  "8️⃣",
		9:  "9️⃣",
		10: "🔟",
	}
	currencies = []string{"AUD", "BRL", "CAD", "CHF", "CNY",
		"CZK", "DKK", "EUR", "GBP", "HKD", "HUF", "IDR", "ILS", "INR",
		"JPY", "KRW", "MXN", "MYR", "NOK", "NZD", "PHP", "PLN", "RUB",
		"SEK", "SGD", "THB", "TRY", "ZAR", "USD"}
	rates = make(map[string]float32)
)

func main() {
	initLog()
	initConfig()
	initChatbase()
	initStrings()
	initKeyboard()
	initDB()
	initTwitter()
	initCache()
	initListings()
	initRates()

	var err error
	bot, err = tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	go checkTransactions()
	go checkTwitter()
	go checkPrice()
	go checkPeriodsPrice()
	//go weeklyRoundUp()
	go checkPosts()
	go updateRates()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s ", update.Message.From.FirstName, update.Message.Text)
			if update.Message.IsCommand() {
				if int64(update.Message.From.ID) != update.Message.Chat.ID {
					//disable commands for groups
					continue
				}
				command := update.Message.Command()
				sendMetric(update.Message.From.ID, command, update.Message.Text)
				switch command {
				case "start":
					go start(update.Message)
				case "addwallet":
					addWallet(update.Message)
				case "resetwallets":
					resetWallets(update.Message)
				case "balance":
					balance(update.Message)
				case "index":
					index(update.Message)
				case "xrp", "price", "p":
					price(update.Message)
				case "chart":
					chart(update.Message)
				case "stats":
					stats(update.Message)
				case "currency":
					currency(update.Message)
				case "newpost":
					newPost(update.Message)
				case "deletepost":
					deletePost(update.Message)
				case "pendingposts":
					pendingPosts(update.Message)
				}
			} else if containsInt64(config.AdminIds, update.Message.Chat.ID) {
				switch currState {
				case "waitingForPost":
					rememberPost(update.Message)
				case "waitingForDelay":
					rememberDelay(update.Message)
				case "waitingForDestination":
					rememberDestination(update.Message)
				}
			}
		} else if update.CallbackQuery != nil {
			sendMetric(update.CallbackQuery.Message.From.ID, update.CallbackQuery.Data, update.CallbackQuery.Data)
			m := update.CallbackQuery.Message
			m.Text = update.CallbackQuery.Data
			m.From = update.CallbackQuery.From
			switch update.CallbackQuery.Data {
			case "stats":
				stats(m)
			case "start":
				go start(m)
			case "help":
				help(m)
			case "index":
				index(m)
			case "balance":
				balance(m)
			case "addwallet":
				addWallet(m)
			case "currency":
				currency(m)
			case "chart 30d", "chart 24h", "chart":
				chart(m)
			case "yes":
				resetWalletsYes(m)
			case "no":
				resetWalletsNo(m)
			}
			bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{update.CallbackQuery.ID, "", false, "", 0})
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func initConfig() {
	readJson(&config, "config.json")
}

func initChatbase() {
	metric = chatbase.New(config.MetricToken)
}

func initStrings() {
	readJson(&phrases, "strings.json")
}

func initKeyboard() {
	start1Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("XRP stats", "stats"),
			tgbotapi.NewInlineKeyboardButtonData("Chart", "chart"),
			tgbotapi.NewInlineKeyboardButtonData("Index", "index"),
		),
	)
	start2Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Balance", "balance"),
			tgbotapi.NewInlineKeyboardButtonData("Add wallet", "addwallet"),
		),
	)
	start3Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Currency", "currency"),
			tgbotapi.NewInlineKeyboardButtonURL("Support", config.SupportURL),
			tgbotapi.NewInlineKeyboardButtonData("CMDs", "help"),
		),
	)
	txKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Transaction details", ""),
			tgbotapi.NewInlineKeyboardButtonURL("Buy XRP", config.BuySellXRP),
		),
	)
	priceKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("XRP stats", "stats"),
			tgbotapi.NewInlineKeyboardButtonURL("Buy XRP", config.BuySellXRP),
			tgbotapi.NewInlineKeyboardButtonData("Start", "start"),
		),
	)
	checkPriceKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("XRP stats", "stats"),
			tgbotapi.NewInlineKeyboardButtonURL("Tweet!", ""),
			tgbotapi.NewInlineKeyboardButtonData("Start", "start"),
		),
	)
	balanceKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("XRP stats", "stats"),
			tgbotapi.NewInlineKeyboardButtonURL("Buy XRP", config.BuySellXRP),
			tgbotapi.NewInlineKeyboardButtonData("Start", "start"),
		),
	)
	statsKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Chart", "chart"),
			tgbotapi.NewInlineKeyboardButtonURL("Trade XRP", config.BuySellXRP),
			tgbotapi.NewInlineKeyboardButtonData("Start", "start"),
		),
	)
	indexKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("XRP stats", "stats"),
			tgbotapi.NewInlineKeyboardButtonData("Chart", "chart"),
			tgbotapi.NewInlineKeyboardButtonData("Start", "start"),
		),
	)
	chart24hKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Chart 30d", "chart 30d"),
			tgbotapi.NewInlineKeyboardButtonURL("Buy XRP", config.BuySellXRP),
			tgbotapi.NewInlineKeyboardButtonData("Start", "start"),
		),
	)
	chart30dKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Chart 24h", "chart 24h"),
			tgbotapi.NewInlineKeyboardButtonURL("Buy XRP", config.BuySellXRP),
			tgbotapi.NewInlineKeyboardButtonData("Start", "start"),
		),
	)
	yesNoKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", "yes"),
			tgbotapi.NewInlineKeyboardButtonData("No", "no"),
		),
	)
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
	db.AutoMigrate(&User{}, &Wallet{}, &UserWallet{})
	log.Print("Migrated")
}

func initTwitter() {
	twitter = anaconda.NewTwitterApiWithCredentials(config.TwitterAccessToken,
		config.TwitterAccessSecret,
		config.TwitterConsumerKey,
		config.TwitterConsumerSecret)

	v := url.Values{}
	v.Add("exclude_replies", "true")
	v.Add("include_rts", "false")
	for _, val := range config.TwitterAccounts {
		v.Add("screen_name", val)
		v.Add("count", "1")
		searchResult, err := twitter.GetUserTimeline(v)
		if err != nil {
			log.Print(err)
		}
		if len(searchResult) > 0 {
			sinceTwitter[val] = searchResult[0].Id
		}
	}
}

func initCache() {
	loadChart("thirtyMin", "USD")
	loadChart("day", "USD")
	cache24h.Time = time.Now()
}

func initListings() {
	var err error
	listings, err = cmc.Listings()
	if err != nil {
		log.Print(err)
	}
}

func initRates() {
	fixer = fixerio.New()
	fixer.Base("USD")
	var symbols []string
	for _, s := range currencies {
		symbols = append(symbols, s)
	}
	fixer.Symbols(symbols...)

	resp, err := fixer.GetRates()
	if err != nil {
		log.Print(err)
	}
	for k, v := range resp {
		rates[k] = v
	}
}
