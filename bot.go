package main

import (
	"telegram-bot-api"
	"log"
	"os"
	"io"
	"io/ioutil"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"time"
	"github.com/ChimeraCoder/anaconda"
	"net/url"
	"net/http"
	"github.com/m90/go-chatbase"
)

var (
	bot          *tgbotapi.BotAPI
	config       Configuration
	db           *gorm.DB
	metric       *chatbase.Client
	phrases      map[int]string
	cache24h     CachedStats
	cache30d     CachedStats
	sinceTwitter = make(map[string]int64)
	twitter      *anaconda.TwitterApi
	listings     []Listing
	currState    string
	currPost     PendingPost

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
	currencies = []string{"AUD", "BRL", "CAD", "CHF", "CLP", "CNY",
		"CZK", "DKK", "EUR", "GBP", "HKD", "HUF", "IDR", "ILS", "INR",
		"JPY", "KRW", "MXN", "MYR", "NOK", "NZD", "PHP", "PKR", "PLN", "RUB",
		"SEK", "SGD", "THB", "TRY", "TWD", "ZAR", "USD"}
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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s ", update.Message.From.FirstName, update.Message.Text)
			if update.Message.IsCommand() {
				if int64(update.Message.From.ID) != update.Message.Chat.ID{
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
			tgbotapi.NewInlineKeyboardButtonURL("Trade XRP", config.BuySellXRP),
		),
	)
	priceKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("XRP stats", "stats"),
			tgbotapi.NewInlineKeyboardButtonURL("Trade XRP", config.BuySellXRP),
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
			tgbotapi.NewInlineKeyboardButtonURL("Trade XRP", config.BuySellXRP),
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
			tgbotapi.NewInlineKeyboardButtonURL("Trade XRP", config.BuySellXRP),
			tgbotapi.NewInlineKeyboardButtonData("Start", "start"),
		),
	)
	chart30dKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Chart 24h", "chart 24h"),
			tgbotapi.NewInlineKeyboardButtonURL("Trade XRP", config.BuySellXRP),
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
	loadChart("thirtyMin")
	loadChart("day")
	cache24h = CachedStats{Time: time.Now(), Stats: getRippleStats()}
}

func initListings() {
	resp, err := http.Get(config.CoinMarketCapListings)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	str := json.Get(bodyBytes, "data").ToString()
	json.UnmarshalFromString(str, &listings)
}
