package main

import (
	"telegram-bot-api"
	"strings"
	cmc "github.com/coincircle/go-coinmarketcap"
	"github.com/coincircle/go-coinmarketcap/types"
	"log"
	"fmt"
	"time"
)

func start(message *tgbotapi.Message) {
	addUserIfAbsent(message.From)
	var username string
	if message.From.UserName != "" {
		username = "@" + message.From.UserName
	}
	sendMessage(message.Chat.ID, fmt.Sprintf(phrases[28], username), nil)
	time.Sleep(time.Duration(500) * time.Millisecond)
	sendMessage(message.Chat.ID, phrases[25], start1Keyboard)
	time.Sleep(time.Duration(500) * time.Millisecond)
	sendMessage(message.Chat.ID, phrases[26], start2Keyboard)
	time.Sleep(time.Duration(500) * time.Millisecond)
	sendMessage(message.Chat.ID, phrases[27], start3Keyboard)
}

func help(message *tgbotapi.Message) {
	sendMessage(message.Chat.ID, phrases[0], nil)
}

func addWallet(message *tgbotapi.Message) {
	fields := strings.Fields(message.Text)
	if len(fields) == 3 {
		if !checkAddress(fields[1]) {
			sendMessage(message.Chat.ID, phrases[4], nil)
			return
		}
		addWalletDB(message)
		var text string
		text = fmt.Sprintf(phrases[1], fields[2])
		sendMessage(message.Chat.ID, text, nil)
	} else {
		sendMessage(message.Chat.ID, phrases[3], nil)
	}

}

func resetWallets(message *tgbotapi.Message) {
	sendMessage(message.Chat.ID, phrases[21], yesNoKeyboard)
}
func resetWalletsYes(message *tgbotapi.Message) {
	resetWalletsDB(message)
	sendMessage(message.Chat.ID, phrases[2], nil)
}
func resetWalletsNo(message *tgbotapi.Message) {
	sendMessage(message.Chat.ID, phrases[22], nil)
}

func balance(message *tgbotapi.Message) {
	var wallets []Wallet
	user := getUser(message.From.ID)
	db.Model(&user).Association("Wallets").Find(&wallets)
	if len(wallets) == 0 {
		sendMessage(message.Chat.ID, phrases[13], nil)
		return
	}
	balances := make(map[string]float64)
	var sum float64
	for _, wallet := range wallets {
		var uw UserWallet
		db.Find(&uw, "user_id = ? AND wallet_id = ?", user.ID, wallet.ID)
		bal := getBalance(wallet.Address)
		sum += bal
		balances[uw.Name] = bal
	}
	text := fmt.Sprintf(phrases[14], float64ToString(sum))
	for name, bal := range balances {
		text += fmt.Sprintf(phrases[15], name, float64ToString(bal))
	}
	price, err := cmc.Price(&cmc.PriceOptions{
		Symbol:  "XRP",
		Convert: user.Currency,
	})
	if err != nil {
		log.Print(err)
	}
	text += fmt.Sprintf(phrases[16], float64ToString(price*sum), user.Currency)
	sendMessage(message.Chat.ID, text, balanceKeyboard)
}

func index(message *tgbotapi.Message) {
	user := getUser(message.From.ID)
	text := "*Top 10 Cryptocurrencies*\n_(by Market Cap)_\n"
	coins, err := cmc.Tickers(&cmc.TickersOptions{
		Start:   0,
		Limit:   10,
		Convert: user.Currency,
	})
	if err != nil {
		log.Print(err)
	}
	newMap := make(map[int]*types.Ticker)
	for _, coin := range coins {
		newMap[coin.Rank] = coin
	}

	var currText string
	for i := 0; i < 10; i++ {
		number := i
		number++
		numberStr := numberEmojis[number]
		currText = numberStr + " " + newMap[number].Symbol + " " +
			float64ToString(newMap[number].Quotes[user.Currency].Price) + " " + user.Currency + " _(" +
			float64WithSign(newMap[number].Quotes[user.Currency].PercentChange24H) + "%)_\n"

		if newMap[number].Name == "Ripple" {
			currText = "*" + currText + "*"
			currText = strings.Replace(currText, "_", "", -1)
		}
		text += currText
	}
	sendMessage(message.Chat.ID, text, indexKeyboard)
}

func price(message *tgbotapi.Message) {
	user := getUser(message.From.ID)
	var text string
	fields := strings.Fields(message.Text)
	ticker := "XRP"
	if len(fields) == 2 {
		ticker = getCurrency(fields[1])
	}
	if ticker != "" {
		coin, err := cmc.Ticker(&cmc.TickerOptions{
			Symbol:  ticker,
			Convert: user.Currency,
		})
		if err != nil {
			log.Print(err)
		}

		text = fmt.Sprintf(phrases[17], coin.Symbol,
			float64ToStringPrec3(coin.Quotes[user.Currency].Price),
			user.Currency,
			float64WithSign(coin.Quotes[user.Currency].PercentChange24H),
			float64WithSign(coin.Quotes[user.Currency].PercentChange7D))

	} else {
		text = phrases[5]
	}

	if ticker == "XRP" {
		sendMessage(message.Chat.ID, text, priceKeyboard)
	} else {
		sendMessage(message.Chat.ID, text, nil)
	}
}

func chart(message *tgbotapi.Message) {
	fields := strings.Fields(message.Text)
	if len(fields) == 1 {
		chart24h(message)
	} else if len(fields) == 2 {
		if fields[1] == "24h" {
			chart24h(message)
		} else if fields[1] == "30d" {
			chart30d(message)
		}
	} else {
		sendMessage(message.Chat.ID, phrases[8], nil)
	}

}

func chart24h(message *tgbotapi.Message) {
	var photo tgbotapi.PhotoConfig
	setUploadingPhoto(message.Chat.ID)
	user := getUser(message.From.ID)
	coin := getRippleStats(user.Currency)
	if time.Now().Sub(cache24h.Time).Minutes() > 3 || cache24h.PhotoIds[user.Currency] == "" {
		loadChart("thirtyMin", user.Currency)
		cache24h.Time = time.Now()
		photo = tgbotapi.NewPhotoUpload(message.Chat.ID, "./charts/chart-thirtyMin" + user.Currency + ".png")
		photo.Caption = "*XRP* (24h) | " + "*Price:* " +
			float64ToStringPrec3(coin.Quotes[user.Currency].Price) + " " + user.Currency
		photo.ParseMode = tgbotapi.ModeMarkdown
		photo.BaseChat.ReplyMarkup = &chart24hKeyboard
		if id := sendPhoto(photo); id != "" {
			cache24h.PhotoIds[user.Currency] = id
		}
	} else {
		photo = tgbotapi.NewPhotoShare(message.Chat.ID, cache24h.PhotoIds[user.Currency])
		photo.Caption = "*XRP* (24h) | " + "*Price:* " +
			float64ToStringPrec3(coin.Quotes[user.Currency].Price) + " " + user.Currency
		photo.ParseMode = tgbotapi.ModeMarkdown
		photo.BaseChat.ReplyMarkup = &chart24hKeyboard
		sendPhoto(photo)
	}
}

func chart30d(message *tgbotapi.Message) {
	var photo tgbotapi.PhotoConfig
	setUploadingPhoto(message.Chat.ID)
	user := getUser(message.From.ID)
	coin := getRippleStats(user.Currency)
	if time.Now().Sub(cache30d.Time).Minutes() > 3 || cache30d.PhotoIds[user.Currency] == "" {
		loadChart("day", user.Currency)
		cache30d.Time = time.Now()
		photo = tgbotapi.NewPhotoUpload(message.Chat.ID, "./charts/chart-day" + user.Currency + ".png")
		photo.Caption = "*XRP* (30d) | " + "*Price:* " +
			float64ToStringPrec3(coin.Quotes[user.Currency].Price) + " " + user.Currency
		photo.ParseMode = tgbotapi.ModeMarkdown
		photo.BaseChat.ReplyMarkup = &chart30dKeyboard
		if id := sendPhoto(photo); id != "" {
			cache30d.PhotoIds[user.Currency] = id
		}
	} else {
		photo = tgbotapi.NewPhotoShare(message.Chat.ID, cache30d.PhotoIds[user.Currency])
		photo.Caption = "*XRP* (30d) | " + "*Price:* " +
			float64ToStringPrec3(coin.Quotes[user.Currency].Price) + " " + user.Currency
		photo.ParseMode = tgbotapi.ModeMarkdown
		photo.BaseChat.ReplyMarkup = &chart30dKeyboard
		sendPhoto(photo)
	}
}

func stats(message *tgbotapi.Message) {
	user := getUser(message.From.ID)
	coin, err := cmc.Ticker(&cmc.TickerOptions{
		Symbol:  "XRP",
		Convert: user.Currency,
	})
	if err != nil {
		log.Print(err)
	}
	market, err := cmc.GlobalMarket(&cmc.GlobalMarketOptions{
		Convert: user.Currency,
	})
	if err != nil {
		log.Print(err)
	}

	share := coin.Quotes[user.Currency].MarketCap * 100 / market.Quotes[user.Currency].TotalMarketCap

	text := fmt.Sprintf(phrases[18], float64ToString(coin.Quotes[user.Currency].Price),
		user.Currency,
		float64WithSign(coin.Quotes[user.Currency].PercentChange24H),
		float64ToString(coin.Quotes[user.Currency].Volume24H/1000000),
		user.Currency,
		float64ToString(coin.Quotes[user.Currency].MarketCap/1000000000),
		user.Currency,
		float64ToString(share),
		float64ToString(market.BitcoinPercentageOfMarketCap))
	sendMessage(message.Chat.ID, text, statsKeyboard)
}

func currency(message *tgbotapi.Message) {
	fields := strings.Fields(message.Text)
	if len(fields) < 2 {
		sendMessage(message.Chat.ID, phrases[7], nil)
		return
	}
	currency := strings.ToUpper(fields[1])
	if !contains(currencies, currency) && currency != "EURO"{
		sendMessage(message.Chat.ID, phrases[5], nil)
		return
	}
	user := getUser(message.From.ID)
	if currency == "EURO"{
		user.Currency = "EUR"
	}else {
		user.Currency = currency
	}
	db.Save(&user)
	text := fmt.Sprintf(phrases[6], user.Currency)
	sendMessage(message.Chat.ID, text, nil)
}

func newPost(message *tgbotapi.Message) {
	if !containsInt64(config.AdminIds, message.Chat.ID) {
		return
	}
	currState = "waitingForPost"
	sendMessage(message.Chat.ID, phrases[9], nil)
}

func deletePost(message *tgbotapi.Message) {
	if !containsInt64(config.AdminIds, message.Chat.ID) {
		return
	}
	var posts []PendingPost
	readJson(&posts, "posts.json")
	fields := strings.Fields(message.Text)
	ind := stringToInt64(fields[1])
	for i := range posts {
		if int64(i) == ind-1 {
			posts = append(posts[:i], posts[i+1:]...)
			break
		}
	}
	writeJson(&posts, "posts.json")
	sendMessage(message.Chat.ID, phrases[12], nil)
}

func pendingPosts(message *tgbotapi.Message) {
	if !containsInt64(config.AdminIds, message.Chat.ID) {
		return
	}
	var posts []PendingPost
	readJson(&posts, "posts.json")
	var text string
	for i, post := range posts {
		text += int64ToString(int64(i+1)) + ". " + post.Message.Text + "\n"
	}
	sendMessage(message.Chat.ID, text, nil)
}
