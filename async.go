package main

import (
	"log"
	"time"
	"net/url"
	cmc "github.com/coincircle/go-coinmarketcap"
	"fmt"
	"telegram-bot-api"
	"strings"
)

func checkTransactions() {
	t := time.Now().UTC().Format(time.RFC3339)
	for {
		time.Sleep(1 * time.Minute)
		log.Print("Started checking")
		rows, _ := db.Table("wallets").Rows()
		for rows.Next() {
			var wallet Wallet
			db.ScanRows(rows, &wallet)
			txs := getTransactions(wallet.Address, t)
			if len(txs) > 0 {
				go sendNotifications(txs, wallet)
			}
		}
		t = time.Now().UTC().Format(time.RFC3339)
	}
}

func checkTwitter() {
	for {
		v := url.Values{}
		v.Add("exclude_replies", "true")
		v.Add("include_rts", "false")
		for _, val := range config.TwitterAccounts {
			v.Add("screen_name", val)
			since := int64ToString(sinceTwitter[val])
			v.Add("since_id", since)
			searchResult, err := twitter.GetUserTimeline(v)
			if err != nil {
				log.Print(err)
			}
			for _, val := range searchResult {
				text := val.User.Name+"("+val.User.ScreenName+"):\n"+val.FullText
				sendMessage(config.ChannelId, text, nil)
				sendMessage(config.ChatId, text, nil)
			}
			if len(searchResult) > 0 {
				sinceTwitter[val] = searchResult[0].Id
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func checkPrice() {
	for {
		var t TimeForSending
		readJson(&t, "time.json")
		if t.anyTime() {
			var old Prices
			readJson(&old, "prices.json")
			coin, err := cmc.Ticker(&cmc.TickerOptions{
				Symbol:  "XRP",
				Convert: "USD",
			})
			if err != nil {
				log.Print(err)
			}

			var textTemplate string
			if coin.Quotes["USD"].PercentChange24H >= 0 {
				textTemplate = phrases[23]
			} else {
				textTemplate = phrases[24]
			}

			text := fmt.Sprintf(textTemplate,
				float64WithSign(coin.Quotes["USD"].PercentChange24H),
				float64ToStringPrec3(coin.Quotes["USD"].Price))

			textForUrl := strings.Replace(
				strings.Replace(text, "*", "", -1), " ", " $", 1) +
				" (via @XRPwatch)" + phrases[20]
			myUrl, err := url.Parse(config.TwitterShareURL)
			if err != nil {
				log.Print(err)
			}
			parameters := url.Values{}
			parameters.Add("text", textForUrl)
			myUrl.RawQuery = parameters.Encode()
			urlStr := myUrl.String()
			keyboard := checkPriceKeyboard
			keyboard.InlineKeyboard[0][1].URL = &urlStr

			if time.Now().After(t.ChannelTime) {
				sendMessage(config.ChannelId, text, nil)
				t.ChannelTime =
					t.ChannelTime.Add(time.Duration(config.ChannelHours) * time.Hour)
			}
			if time.Now().After(t.GroupTime) {
				sendMessage(config.ChatId, text, nil)
				t.GroupTime =
					t.GroupTime.Add(time.Duration(config.GroupHours) * time.Hour)
			}
			if time.Now().After(t.UsersTime) {
				userText := fmt.Sprintf(textTemplate,
					float64WithSign(coin.Quotes["USD"].PercentChange24H),
					"%v")
				userText = strings.Replace(userText, "%", "%%", 1)
				go convertAndSendAllUsers(userText, coin.Quotes["USD"].Price, keyboard)
				t.UsersTime =
					t.UsersTime.Add(time.Duration(config.UsersHours) * time.Hour)
			}
			if time.Now().After(t.TwitterTime) {
				text = strings.Replace(text, "*", "", -1)
				text = strings.Replace(text, " ", " $", 1)
				tweet(text)
				t.TwitterTime =
					t.TwitterTime.Add(time.Duration(config.TwitterHours) * time.Hour)
			}
			writeJson(&old, "prices.json")
			writeJson(&t, "time.json")
		}
		time.Sleep(1 * time.Hour)
	}
}

func checkPeriodsPrice() {
	for {
		var old Prices
		readJson(&old, "prices.json")

		price, err := cmc.Price(&cmc.PriceOptions{
			Symbol:  "XRP",
			Convert: "USD",
		})
		if err != nil {
			log.Print(err)
		}

		templateHigh := "🚀 XRP on a new *%v high* @ %v USD"
		templateLow := "📉 XRP on a new *%v low* @ %v USD"
		var textWOprice, textFinal string
		//HIGHS CHECK
		allTimeHigh := true
		threeMonthsHigh := true
		monthHigh := true
		weekHigh := true

		if old.Highs.AllTime > price {
			allTimeHigh = false
		}
		if allTimeHigh == false {
			for _, val := range old.Highs.ThreeMonths {
				if val > price {
					threeMonthsHigh = false
				}
			}
			if threeMonthsHigh == false {
				for _, val := range old.Highs.Month {
					if val > price {
						monthHigh = false
					}
				}
				if monthHigh == false {
					for _, val := range old.Highs.Week {
						if val > price {
							weekHigh = false
						}
					}
					if weekHigh == true {
						//post week
						textWOprice = fmt.Sprintf(templateHigh, "7d", "%v")
					}
				} else {
					//post month
					textWOprice = fmt.Sprintf(templateHigh, "30d", "%v")
				}
			} else {
				//post 3m
				textWOprice = fmt.Sprintf(templateHigh, "3m", "%v")
			}
		} else {
			//post allTime
			textWOprice = fmt.Sprintf(templateHigh, "all-timr", "%v")
		}
		textFinal += fmt.Sprintf(textWOprice, float64ToStringPrec3(price))

		if !allTimeHigh && !threeMonthsHigh && !monthHigh && !weekHigh {
			//LOWS CHECK
			threeMonthsLow := true
			monthLow := true
			weekLow := true
			for _, val := range old.Lows.ThreeMonths {
				if val < price {
					threeMonthsLow = false
				}
			}

			if threeMonthsLow == false {
				for _, val := range old.Lows.Month {
					if val < price {
						monthLow = false
					}
				}
				if monthLow == false {
					for _, val := range old.Lows.Week {
						if val < price {
							weekLow = false
						}
					}
					if weekLow == true {
						//post week
						textWOprice = fmt.Sprintf(templateLow, "7d", "%v")
					}
				} else {
					//post month
					textWOprice = fmt.Sprintf(templateLow, "30d", "%v")
				}
			} else {
				//post 3m
				textWOprice = fmt.Sprintf(templateLow, "3m", "%v")
			}
			if weekLow == false {
				textFinal = ""
			} else {
				textFinal += fmt.Sprintf(textWOprice, float64ToStringPrec3(price))
			}
		}
		if textFinal != "" {
			sendMessage(config.ChannelId, textFinal, nil)
			sendMessage(config.ChatId, textFinal, nil)
			convertAndSendAllUsers(textWOprice, price, nil)
			textFinal = strings.Replace(textFinal, "*", "", -1)
			textFinal = strings.Replace(textFinal, " ", " $", 1)
			tweet(textFinal)
		}

		if old.Highs.AllTime < price {
			old.Highs.AllTime = price
		}
		//if day changed
		oldTime, err := time.Parse(time.RFC822, old.LastCheck)
		if err != nil {
			log.Print(err)
		}
		if oldTime.Day() != time.Now().Day() {
			shiftArray(&old.Highs.ThreeMonths)
			shiftArray(&old.Highs.Month)
			shiftArray(&old.Highs.Week)
			shiftArray(&old.Lows.ThreeMonths)
			shiftArray(&old.Lows.Month)
			shiftArray(&old.Lows.Week)
		}
		old.LastCheck = time.Now().Format(time.RFC822)

		if old.Highs.ThreeMonths[len(old.Highs.ThreeMonths)-1] < price ||
			old.Highs.ThreeMonths[len(old.Highs.ThreeMonths)-1] == 0 {
			old.Highs.ThreeMonths[len(old.Highs.ThreeMonths)-1] = price
		}
		if old.Highs.Month[len(old.Highs.Month)-1] < price ||
			old.Highs.Month[len(old.Highs.Month)-1] == 0 {
			old.Highs.Month[len(old.Highs.Month)-1] = price
		}
		if old.Highs.Week[len(old.Highs.Week)-1] < price ||
			old.Highs.Week[len(old.Highs.Week)-1] == 0 {
			old.Highs.Week[len(old.Highs.Week)-1] = price
		}

		if old.Lows.ThreeMonths[len(old.Lows.ThreeMonths)-1] > price ||
			old.Lows.ThreeMonths[len(old.Lows.ThreeMonths)-1] == 0 {
			old.Lows.ThreeMonths[len(old.Lows.ThreeMonths)-1] = price
		}
		if old.Lows.Month[len(old.Lows.Month)-1] > price ||
			old.Lows.Month[len(old.Lows.Month)-1] == 0 {
			old.Lows.Month[len(old.Lows.Month)-1] = price
		}
		if old.Lows.Week[len(old.Lows.Week)-1] > price ||
			old.Lows.Week[len(old.Lows.Week)-1] == 0 {
			old.Lows.Week[len(old.Lows.Week)-1] = price
		}

		writeJson(&old, "prices.json")
		time.Sleep(15 * time.Minute)
	}
}

func weeklyRoundUp() {
	for {
		if time.Now().Weekday() == time.Sunday && time.Now().Hour() == 10 {
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

			text := "Weekly roundup:\nXRP's current price is %s USD with a total " +
				"market cap of %sbn USD. XRP market share " +
				"went to %s%% and BTC dominance to" +
				"%s%%. 👉 Discuss @XRPchats"

			share := coin.Quotes["USD"].MarketCap * 100 / market.Quotes["USD"].TotalMarketCap

			text = fmt.Sprintf(text,
				float64ToStringPrec3(coin.Quotes["USD"].Price),
				float64ToString(coin.Quotes["USD"].MarketCap/1000000000),
				float64ToString(share),
				float64ToString(market.BitcoinPercentageOfMarketCap))
			sendMessage(config.ChannelId, text, nil)
			sendMessage(config.ChatId, text, nil)
			//sendAllUsers(tgbotapi.MessageConfig{Text: text})
		}
		time.Sleep(1 * time.Hour)
	}
}

func checkPosts() {
	for {
		time.Sleep(1 * time.Minute)
		var posts []PendingPost
		readJson(&posts, "posts.json")

		sent := false
		for i, post := range posts {
			t1 := time.Now().UTC()
			t2 := post.PostTime.UTC()
			d := t1.Sub(t2).Seconds()
			if d > 0 {
				sent = true
				if post.Destination == 1 {
					msg := parsePost(&posts[i], config.ChannelId)
					bot.Send(msg)
				} else if post.Destination == 2 {
					msg := parsePost(&posts[i], config.ChatId)
					bot.Send(msg)
				} else {
					sendAllUsersMessageConfig(parsePost(&posts[i], 0).(tgbotapi.MessageConfig))
				}
			}
		}
		if sent {
			writeJson(&posts, "posts.json")
		}
	}
}

func parsePost(p *PendingPost, id int64) tgbotapi.Chattable {
	var msg tgbotapi.Chattable

	if p.Message.Text != "" {
		msg = tgbotapi.NewMessage(id, p.Message.Text)
	} else if p.Message.Video != nil {
		msg = tgbotapi.NewVideoShare(id, p.Message.Video.FileID)
		msg2 := msg.(tgbotapi.VideoConfig)
		msg2.Caption = p.Message.Caption
		msg = msg2
	} else if p.Message.Sticker != nil {
		msg = tgbotapi.NewStickerShare(id, p.Message.Sticker.FileID)
	} else if p.Message.Photo != nil {
		msg = tgbotapi.NewPhotoShare(id, (*p.Message.Photo)[0].FileID)
		msg2 := msg.(tgbotapi.PhotoConfig)
		msg2.Caption = p.Message.Caption
		msg = msg2
	} else if p.Message.Document != nil {
		msg = tgbotapi.NewDocumentShare(id, p.Message.Document.FileID)
		msg2 := msg.(tgbotapi.DocumentConfig)
		msg2.Caption = p.Message.Caption
		msg = msg2
	} else if p.Message.Audio != nil {
		msg = tgbotapi.NewAudioShare(id, p.Message.Audio.FileID)
		msg2 := msg.(tgbotapi.AudioConfig)
		msg2.Caption = p.Message.Caption
		msg = msg2
	} else if p.Message.VideoNote != nil {
		msg = tgbotapi.NewVideoNoteShare(id, p.Message.VideoNote.Length,
			p.Message.VideoNote.FileID)
	} else if p.Message.Voice != nil {
		msg = tgbotapi.NewVoiceShare(id, p.Message.Voice.FileID)
	}
	p.PostTime = time.Now().Add(time.Duration(p.DelayHours) * time.Hour)
	return msg
}

func updateRates(){
	for{
		time.Sleep(24*time.Hour)
		log.Print("Updating rates")
		rates = make(map[string]float64)
		var symbols []string
		for _, s := range currencies {
			symbols = append(symbols, s)
		}
		fixer.Symbols(symbols...)

		resp, err := fixer.GetRates()
		if err != nil {
			log.Print(err)
		}
		for k, v := range resp{
			rates[k] = float64(v)
		}
		log.Print(rates)
	}
}