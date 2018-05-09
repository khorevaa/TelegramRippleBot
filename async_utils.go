package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"github.com/json-iterator/go"
	"github.com/ChimeraCoder/anaconda"
	"strconv"
	cmc "github.com/coincircle/go-coinmarketcap"

)

var(
	 json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func getTransactions(address string, timestamp string) []Transaction {
	url := configuration.RippleUrlBase + address + configuration.RippleUrlParams + timestamp
	log.Print(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		log.Print(err)
	}
	log.Print(string(bodyBytes))
	var txs []Transaction
	str := json.Get(bodyBytes, "transactions").ToString()
	log.Print(str)
	json.UnmarshalFromString(str, &txs)

	return txs
}

func sendNotifications(txs []Transaction, wallet Wallet){
	var users []User
	db.Model(&wallet).Related(&users, "Users")
	for _, user := range users{
		var text string
		for _, tx := range txs{
			amount, err := strconv.ParseInt(tx.Tx.Amount, 10, 64)
			if err != nil{
				log.Print(err)
			}
			decAmount := float64(amount) / 1000000
			decAmountStr := strconv.FormatFloat(decAmount, 'f', 0, 64)

			var uw UserWallet
			db.First(&uw, "user_id = ? AND wallet_id = ?", user.ID, wallet.ID)
			name := "your"
			if uw.Name != ""{
				name = "*" + uw.Name + "*"
			}

			var balance string
			for _, node := range tx.Meta.AffectedNodes{
				if node.Modified.Data.Account == wallet.Address{
					balance = node.Modified.Data.Balance
				}
			}
			balanceInt, err := strconv.ParseInt(balance, 10, 64)
			if err != nil{
				log.Print(err)
			}
			decBalance := float64(balanceInt) / 1000000
			decBalanceStr := strconv.FormatFloat(decBalance, 'f', 3, 64)


			if tx.Tx.Destination == wallet.Address{
				text = "💰 You received *" + decAmountStr + " XRP* on "
			}else {
				text = "💸 You sent *" + decAmountStr + " XRP* from "
			}
			text +=  name + " wallet\n\n" + "New balance:\n*" + decBalanceStr + " XRP* ≈ "
			coin, err := cmc.GetCoinData("ripple")
			if err != nil {
				log.Print(err)
			}
			text += float64ToString(coin.PriceUSD * decBalance) + " USD\n"
			*linksKeyboard.InlineKeyboard[0][0].URL = "https://xrpcharts.ripple.com/#/transactions/" + tx.Hash
			sendMessage(user.ID, text, linksKeyboard)
		}

	}
}

func postTweet(t anaconda.Tweet){
	sendMessage(configuration.ChannelId, t.User.Name + "(" + t.User.ScreenName + "):\n" + t.FullText, nil)
}