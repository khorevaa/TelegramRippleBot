package main

import (
	"log"
	"time"
	"net/url"
	cmc "github.com/coincircle/go-coinmarketcap"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
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
			var since string
			if sinceTwitter[val] == 0 {
				sinceTwitter[val] = 1
				since = "1"
			} else {
				since = int64ToString(sinceTwitter[val])
			}
			v.Add("since_id", since)
			searchResult, err := twitter.GetUserTimeline(v)
			if err != nil {
				log.Print(err)
			}
			for _, val := range searchResult {
				postTweet(val)
			}
			if len(searchResult) > 0 {
				sinceTwitter[val] = searchResult[0].Id
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func checkEverydayPrice() {
	for {
		if time.Now().Hour() == 10 && time.Now().Minute() == 0 {
			var old Prices
			readJson(&old, "prices.json")

			price, err := cmc.Price(&cmc.PriceOptions{
				Symbol:  "XRP",
				Convert: "USD",
			})
			if err != nil {
				log.Print(err)
			}

			var text string
			diff := price - old.Yesterday
			if diff >= 0 {
				text = "🚀 XRP is up %s%% the last 24h and is now trading @ %s USD."
			} else {
				text = "📉 XRP is down %s%% the last 24h and is now trading @ %s USD."
			}

			text = fmt.Sprintf(text, float64WithSign(diff), float64ToString(price))

			sendMessage(config.ChannelId, text, nil)
			tweet(text)
			old.Yesterday = price
			writeJson(&old, "prices.json")
			time.Sleep(24 * time.Hour)
		} else {
			time.Sleep(1 * time.Minute)
		}

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

		text := "🚀 XRP on a new "
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
						text += "7d"
					}
				} else {
					//post month
					text += "30d"
				}
			} else {
				//post 3m
				text += "3m"
			}
		} else {
			//post allTime
			text += "all-time"
		}
		text += fmt.Sprintf(" high @ %s USD", float64ToString(price))

		if !allTimeHigh && !threeMonthsHigh && !monthHigh && !weekHigh {
			//LOWS CHECK
			text = "📉 XRP on a new "
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
						text += "7d"
					}
				} else {
					//post month
					text += "30d"
				}
			} else {
				//post 3m
				text += "3m"
			}
			if weekLow == false {
				text = ""
			} else {
				text += fmt.Sprintf(" low @ %s USD", float64ToString(price))
			}
		}
		if text != "" {
			sendMessage(config.ChannelId, text, nil)
			tweet(text)
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

		if old.Highs.ThreeMonths[len(old.Highs.ThreeMonths)-1] < price {
			old.Highs.ThreeMonths[len(old.Highs.ThreeMonths)-1] = price
		}
		if old.Highs.Month[len(old.Highs.Month)-1] < price {
			old.Highs.Month[len(old.Highs.Month)-1] = price
		}
		if old.Highs.Week[len(old.Highs.Week)-1] < price {
			old.Highs.Week[len(old.Highs.Week)-1] = price
		}

		if old.Lows.ThreeMonths[len(old.Lows.ThreeMonths)-1] > price {
			old.Lows.ThreeMonths[len(old.Lows.ThreeMonths)-1] = price
		}
		if old.Lows.Month[len(old.Lows.Month)-1] > price {
			old.Lows.Month[len(old.Lows.Month)-1] = price
		}
		if old.Lows.Week[len(old.Lows.Week)-1] > price {
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
				float64ToString(coin.Quotes["USD"].Price),
				float64ToString(coin.Quotes["USD"].MarketCap/1000000000),
				float64ToString(share),
				float64ToString(market.BitcoinPercentageOfMarketCap))
			sendMessage(config.ChannelId, text, nil)
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
				sendPost(&posts[i])
			}
		}
		if sent {
			writeJson(&posts, "posts.json")
		}
	}
}

func sendPost(p *PendingPost) {
	var msg tgbotapi.Chattable
	if p.Message.Text != "" {
		msg = tgbotapi.NewMessage(config.ChannelId, p.Message.Text)
	} else if p.Message.Video != nil {
		msg = tgbotapi.NewVideoShare(config.ChannelId, p.Message.Video.FileID)
		msg2 := msg.(tgbotapi.VideoConfig)
		msg2.Caption = p.Message.Caption
		msg = msg2
	} else if p.Message.Sticker != nil {
		msg = tgbotapi.NewStickerShare(config.ChannelId, p.Message.Sticker.FileID)
	} else if p.Message.Photo != nil {
		msg = tgbotapi.NewPhotoShare(config.ChannelId, (*p.Message.Photo)[0].FileID)
		msg2 := msg.(tgbotapi.PhotoConfig)
		msg2.Caption = p.Message.Caption
		msg = msg2
	} else if p.Message.Document != nil {
		msg = tgbotapi.NewDocumentShare(config.ChannelId, p.Message.Document.FileID)
		msg2 := msg.(tgbotapi.DocumentConfig)
		msg2.Caption = p.Message.Caption
		msg = msg2
	} else if p.Message.Audio != nil {
		msg = tgbotapi.NewAudioShare(config.ChannelId, p.Message.Audio.FileID)
		msg2 := msg.(tgbotapi.AudioConfig)
		msg2.Caption = p.Message.Caption
		msg = msg2
	} else if p.Message.VideoNote != nil {
		msg = tgbotapi.NewVideoNoteShare(config.ChannelId, p.Message.VideoNote.Length, p.Message.VideoNote.FileID)
	} else if p.Message.Voice != nil {
		msg = tgbotapi.NewVoiceShare(config.ChannelId, p.Message.Voice.FileID)
	}
	bot.Send(msg)
	p.PostTime = p.PostTime.Add(time.Duration(p.DelayHours) * time.Hour)
}
