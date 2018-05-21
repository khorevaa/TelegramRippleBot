package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"github.com/json-iterator/go"
	cmc "github.com/coincircle/go-coinmarketcap"

	"net/url"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func getTransactions(address string, timestamp string) []Transaction {
	url := config.RippleUrlBase + address + config.RippleUrlParams + timestamp
	log.Print(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}
	//log.Print(string(bodyBytes))
	var txs []Transaction
	str := json.Get(bodyBytes, "transactions").ToString()
	log.Print(str)
	json.UnmarshalFromString(str, &txs)

	return txs
}

func sendNotifications(txs []Transaction, wallet Wallet) {
	var users []User
	db.Model(&wallet).Related(&users, "Users")
	for _, user := range users {
		var text string
		for _, tx := range txs {
			amount := stringToInt64(tx.Tx.Amount)

			decAmount := float64(amount) / 1000000
			decAmountStr := float64ToString(decAmount)

			var uw UserWallet
			db.First(&uw, "user_id = ? AND wallet_id = ?", user.ID, wallet.ID)
			name := "your"
			if uw.Name != "" {
				name = "*" + uw.Name + "*"
			}

			var balance string
			for _, node := range tx.Meta.AffectedNodes {
				if node.Modified.Data.Account == wallet.Address {
					balance = node.Modified.Data.Balance
				}
			}
			balanceInt := stringToInt64(balance)

			decBalance := float64(balanceInt) / 1000000
			decBalanceStr := float64ToString(decBalance)

			if tx.Tx.Destination == wallet.Address {
				text = "💰 You received *" + decAmountStr + " XRP* on "
			} else {
				text = "💸 You sent *" + decAmountStr + " XRP* from "
			}
			text += name + " wallet\n\n" + "New balance:\n*" + decBalanceStr + " XRP* ≈ "
			price, err := cmc.Price(&cmc.PriceOptions{
				Symbol:  "XRP",
				Convert: "USD",
			})
			if err != nil {
				log.Print(err)
			}
			text += float64ToString(price*decBalance) + " USD\n"
			*txKeyboard.InlineKeyboard[0][0].URL =
				"https://xrpcharts.ripple.com/#/transactions/" + tx.Hash
			sendMessage(user.ID, text, txKeyboard)
		}

	}
}


func shiftArray(arr *[]float64) {
	for i := range *arr {
		if i+1 <= len(*arr)-1 {
			(*arr)[i] = (*arr)[i+1]
		} else {
			(*arr)[i] = 0
		}
	}
}

func tweet(text string) {
	_, err := twitter.PostTweet(text+phrases[20], url.Values{})
	if err != nil {
		log.Print(err)
	}
}
