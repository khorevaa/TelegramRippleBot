package main

import (
	"log"
	"time"
	"net/url"
	"strconv"
	"io/ioutil"
	"os"
	cmc "github.com/coincircle/go-coinmarketcap"

	"fmt"
)

func checkTransactions() {
	t := time.Now().UTC().Format(time.RFC3339)
	for {
		time.Sleep(1 * time.Minute)
		log.Print("Started checking")
		rows, _ := db.Table("wallets").Rows()
		for rows.Next() {
			log.Print("Next row")
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
		for _, val := range configuration.TwitterAccounts {
			v.Add("screen_name", val)
			var since string
			if sinceTwitter[val] == 0 {
				sinceTwitter[val] = 1
				since = "1"
			} else {
				since = strconv.FormatInt(sinceTwitter[val], 10)
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
			file, err := os.OpenFile("prices.json", os.O_RDWR, 0644)
			if err != nil {
				log.Print(err)
			}
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&old)
			if err != nil {
				log.Print(err)
			}
			file.Close()

			price, err := cmc.GetCoinPriceUSD("ripple")
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

			sendMessage(configuration.ChannelId, text, nil)
			old.Yesterday = price
			dataJson, err := json.Marshal(&old)
			if err != nil {
				log.Print(err)
			}
			ioutil.WriteFile("prices.json", dataJson, 0644)
			time.Sleep(24 * time.Hour)
		} else {
			time.Sleep(1 * time.Minute)
		}

	}
}

func checkPeriodsPrice() {
	for {
		var old Prices
		file, err := os.OpenFile("prices.json", os.O_RDWR, 0644)
		if err != nil {
			log.Print(err)
		}
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&old)
		if err != nil {
			log.Print(err)
		}
		file.Close()

		price, err := cmc.GetCoinPriceUSD("ripple")
		if err != nil {
			log.Print(err)
		}

		text := "🚀 XRP on a new "
		//high check
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
			//low check
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
			if weekLow == false{
				text = ""
			}else {
				text += fmt.Sprintf(" low @ %s USD", float64ToString(price))
			}
		}
		if text != ""{
			sendMessage(configuration.ChannelId, text, nil)
		}

		if old.Highs.AllTime < price {
			old.Highs.AllTime = price
		}
		// если сменился день
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

		dataJson, err := json.Marshal(&old)
		if err != nil {
			log.Print(err)
		}
		ioutil.WriteFile("prices.json", dataJson, 0644)
		time.Sleep(15 * time.Minute)
	}
}

func weeklyRoundUp() {
	for {
		//if time.Now().Weekday() == time.Sunday && time.Now().Hour() == 10{
		if time.Now().Weekday() == time.Friday && time.Now().Hour() == 21{
			coin, err := cmc.GetCoinData("ripple")
			if err != nil {
				log.Print(err)
			}
			market, err := cmc.GetGlobalMarketData()
			if err != nil {
				log.Print(err)
			}

			text := "Weekly roundup:\nXRP's current price is %s USD with a total " +
				"market cap of %sbn USD. XRP market share " +
				"went to %s%% and BTC dominance to" +
				"%s%%. 👉 Discuss @XRPchats"

			share := coin.MarketCapUSD * 100 / market.TotalMarketCapUSD

			text = fmt.Sprintf(text,
				float64ToString(coin.PriceUSD),
				float64ToString(coin.MarketCapUSD/1000000000),
				float64ToString(share),
				float64ToString(market.BitcoinPercentageOfMarketCap))
			sendMessage(configuration.ChannelId, text, nil)
		}
		time.Sleep(1 * time.Hour)
	}
}
