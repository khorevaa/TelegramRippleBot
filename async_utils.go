package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"github.com/json-iterator/go"
	"github.com/ChimeraCoder/anaconda"
	"strconv"
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

func sendNotifications(txs []Transaction, users []User){
	var text string
	for _, val := range txs{
		amount, err := strconv.ParseInt(val.Tx.Amount, 10, 64)
		if err != nil{
			log.Print(err)
		}
		decAmount := float64(amount) / 1000000
		decAmountStr := strconv.FormatFloat(decAmount, 'f', -1, 64)
		text += "New transaction.\nDestination: " + val.Tx.Destination +
			"\nAmount: " + decAmountStr + "XRP\n"
	}
	for _, val := range users{
		sendMessage(val.ID, text, nil)
	}
}

func postTweet(t anaconda.Tweet){
	sendMessage(configuration.ChannelId, t.User.Name + "(" + t.User.ScreenName + "):\n" + t.FullText, nil)
}