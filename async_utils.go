package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"github.com/json-iterator/go"
)

var(
	 json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func getTransactions(address string, timestamp string) []Transaction {
	resp, err := http.Get(configuration.RippleUrlBase + address + configuration.RippleUrlParams + timestamp)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		log.Print(err)
	}
	var txs []Transaction
	str := json.Get(bodyBytes, "transactions").ToString()

	json.UnmarshalFromString(str, &txs)

	return txs
}

func sendNotifications(txs []Transaction, users []User){
	var text string
	for _, val := range txs{
		text += "New transaction.\nDestination: " + val.Tx.Destination +
			"\nAmount: " + val.Tx.Amount.Value + " " + val.Tx.Amount.Currency + "\n"
	}
	for _, val := range users{
		sendMessage(val.ID, text, nil)
	}
}