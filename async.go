package main

import (
	"log"
	"time"
	"net/url"
	"strconv"
)

func checkTransactions() {
	t := time.Now().Format(time.RFC3339)
	for {
		time.Sleep(1 * time.Minute)
		log.Print("Started checking")
		rows, _ := db.Table("wallets").Rows()
		for rows.Next() {
			log.Print("Next row")
			var wallet Wallet
			db.ScanRows(rows, &wallet)
			var users []User
			db.Model(&wallet).Related(&users, "Users")
			txs := getTransactions(wallet.Address, t)
			if len(txs) > 0 {
				go sendNotifications(txs, users)
			}
		}
		t = time.Now().Format(time.RFC3339)
	}
}

func checkTwitter(){
	for{
		v := url.Values{}
		v.Add("exclude_replies", "true")
		v.Add("include_rts", "false")
		for _, val := range configuration.TwitterAccounts{
			v.Add("screen_name", val)
			var since string
			if sinceTwitter[val] == 0{
				sinceTwitter[val] = 1
				since = "1"
			}else {
				since = strconv.FormatInt(sinceTwitter[val], 10)
			}
			v.Add("since_id", since)
			searchResult, err := twitter.GetUserTimeline(v)
			if err != nil{
				log.Print(err)
			}
			for _, val := range searchResult{
				postTweet(val)
			}
			if len(searchResult) > 0{
				sinceTwitter[val] = searchResult[0].Id
			}
		}

		time.Sleep(1*time.Minute)
	}
}