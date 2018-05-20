package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
	cmc "github.com/coincircle/go-coinmarketcap"
	"github.com/coincircle/go-coinmarketcap/types"
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

func balance(message *tgbotapi.Message) {
	var wallets []Wallet
	user := getUser(message.From.ID)
	db.Model(&user).Association("Wallets").Find(&wallets)
	balances := make(map[string]float64)
	var sum float64
	for _, wallet := range wallets {
		var uw UserWallet
		db.Find(&uw, "user_id = ? AND wallet_id = ?", user.ID, wallet.ID)
		bal := getBalance(wallet.Address)
		sum += bal
		balances[uw.Name] = bal
	}
	text := "🏦 You currently hold *" + float64ToString(sum) + " XRP* on your wallets\n\n"
	for name, bal := range balances {
		text += "*" + name + "*: " + float64ToString(bal) + " XRP\n"
	}
	text += "\nEstimated worth:\n"
	price, err := cmc.Price(&cmc.PriceOptions{
		Symbol:  "XRP",
		Convert: "USD",
	})
	if err != nil {
		log.Print(err)
	}
	text += float64ToString(price*sum) + " USD\n"
	text += "👉 [Buy/Sell XRP](" + config.BuySellXRP + ")" + " - XRP /stats"

	sendMessage(message.Chat.ID, text, nil)
}

func index(message *tgbotapi.Message) {
	var text string
	coins, err := cmc.Tickers(&cmc.TickersOptions{
		Start:   0,
		Limit:   10,
		Convert: "USD",
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
			float64ToString(newMap[number].Quotes["USD"].Price) + " USD _(" +
			float64WithSign(newMap[number].Quotes["USD"].PercentChange24H) + "%)_\n"

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
	if len(fields) == 2 {
		priceCoin(message)
	} else {
		priceXrp(message)
	}

}

func priceCoin(message *tgbotapi.Message) {
	var text string
	fields := strings.Fields(message.Text)
	ticker := getCurrency(fields[1])
	if ticker != "" {
		coin, err := cmc.Ticker(&cmc.TickerOptions{
			Symbol:  ticker,
			Convert: "USD",
		})
		if err != nil {
			log.Print(err)
		}

		text = "*" + coin.Symbol + " = " + float64ToString(coin.Quotes["USD"].Price) + " USD* " +
			float64WithSign(coin.Quotes["USD"].PercentChange24H) + "% _(24h)_" + "\n\n" +
			float64WithSign(coin.Quotes["USD"].PercentChange7D) + "% _(7d)_"

	} else {
		text = phrases[5]
	}

	sendMessage(message.Chat.ID, text, nil)
}

func priceXrp(message *tgbotapi.Message) {
	coin, err := cmc.Ticker(&cmc.TickerOptions{
		Symbol:  "XRP",
		Convert: "USD",
	})
	if err != nil {
		log.Print(err)
	}

	text := "*XRP = " + float64ToString(coin.Quotes["USD"].Price) + " USD* " +
		float64WithSign(coin.Quotes["USD"].PercentChange24H) + "% _(24h)_" + "\n\n" +
		float64WithSign(coin.Quotes["USD"].PercentChange7D) + "% _(7d)_" +
		"\n\n📈 view /chart or more XRP /stats \n" + "👉 [Buy/Sell XRP](" + config.BuySellXRP + ")"
	sendMessage(message.Chat.ID, text, nil)
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
	if time.Now().Sub(cache24h.Time).Minutes() > 3 || cache24h.PhotoId == "" {
		loadChart("thirtyMin")
		cache24h.Time = time.Now()
		cache24h.Stats = getRippleStats()
		photo = tgbotapi.NewPhotoUpload(message.Chat.ID, "chart-thirtyMin.png")
		photo.Caption = cache24h.Stats
		if id := sendPhoto(photo); id != "" {
			cache24h.PhotoId = id
		}
	} else {
		photo = tgbotapi.NewPhotoShare(message.Chat.ID, cache24h.PhotoId)
		photo.Caption = cache24h.Stats
		sendPhoto(photo)
	}
}

func chart30d(message *tgbotapi.Message) {
	var photo tgbotapi.PhotoConfig
	setUploadingPhoto(message.Chat.ID)
	if time.Now().Sub(cache30d.Time).Minutes() > 3 || cache30d.PhotoId == "" {
		loadChart("day")
		cache30d.Time = time.Now()
		cache30d.Stats = getRippleStats()
		photo = tgbotapi.NewPhotoUpload(message.Chat.ID, "chart-day.png")
		photo.Caption = cache30d.Stats
		if id := sendPhoto(photo); id != "" {
			cache30d.PhotoId = id
		}
	} else {
		photo = tgbotapi.NewPhotoShare(message.Chat.ID, cache30d.PhotoId)
		photo.Caption = cache30d.Stats
		sendPhoto(photo)
	}
}

func stats(message *tgbotapi.Message) {
	coin, err := cmc.Ticker(&cmc.TickerOptions{
		Symbol:  "XRP",
		Convert: "USD",
	})
	if err != nil {
		log.Print(err)
	}
	market, err := cmc.GlobalMarket(&cmc.GlobalMarketOptions{
		Convert: "USD",
	})
	if err != nil {
		log.Print(err)
	}

	text := "Last price: " + float64ToString(coin.Quotes["USD"].Price) + " USD _(" +
		float64WithSign(coin.Quotes["USD"].PercentChange24H) + "% last 24h)_\n"
	text += "Market cap: " + float64ToString(coin.Quotes["USD"].MarketCap/1000000000) + " bln USD\n"
	share := coin.Quotes["USD"].MarketCap * 100 / market.Quotes["USD"].TotalMarketCap
	text += "Market share: " + float64ToString(share) + "%\n"
	text += "BTC dominance: " + float64ToString(market.BitcoinPercentageOfMarketCap) + "%"

	sendMessage(message.Chat.ID, text, nil)
}

func currency(message *tgbotapi.Message) {
	fields := strings.Fields(message.Text)
	if len(fields) < 2 {
		sendMessage(message.Chat.ID, phrases[7], nil)
		return
	}
	if !contains(currencies, strings.ToUpper(fields[1])) {
		sendMessage(message.Chat.ID, phrases[5], nil)
		return
	}
	user := getUser(message.From.ID)
	user.Currency = strings.ToUpper(fields[1])
	db.Save(&user)
	sendMessage(message.Chat.ID, phrases[6], nil)
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
