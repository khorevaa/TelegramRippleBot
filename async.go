package main

import (
	"log"
	"time"
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
