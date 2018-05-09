package main

import "time"

type Config struct {
	BotToken                                                                           string
	ChannelId                                                                          int64
	BuySellXRP                                                                         string
	TwitterAccounts                                                                    []string
	TwitterAccessToken, TwitterAccessSecret, TwitterConsumerKey, TwitterConsumerSecret string
	RippleUrlBase                                                                      string
	RippleUrlParams                                                                    string
	BittrexChartURL                                                                    string
	CoinMarketCapListings                                                              string
}

type User struct {
	ID        int64    `gorm:"primary_key"`
	FirstName string
	LastName  string
	UserName  string
	Currency  string
	Wallets   []Wallet `gorm:"many2many:user_wallets;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Wallet struct {
	ID        int    `gorm:"primary_key"`
	Address   string
	CreatedAt time.Time
	UpdatedAt time.Time
	Users     []User `gorm:"many2many:user_wallets;"`
}

type UserWallet struct {
	UserID   int64 `gorm:"primary_key"`
	WalletID int64 `gorm:"primary_key"`
	Name     string
}

type Transaction struct {
	Date string   `json:"date"`
	Hash string   `json:"hash"`
	Tx   TxInfo   `json:"tx"`
	Meta MetaInfo `json:"meta"`
}

type TxInfo struct {
	TransactionType, Destination string
	Amount                       string
}

type MetaInfo struct {
	AffectedNodes []Node
}

type Node struct {
	Modified ModifiedNode `json:"ModifiedNode"`
}

type ModifiedNode struct {
	Data FinalFields `json:"FinalFields"`
}

type FinalFields struct {
	Balance, Account string
}

type CachedStats struct {
	Time    time.Time
	Stats   string
	PhotoId string
}

type Candle struct {
	Open    float64 `json:"O"`
	Close   float64 `json:"C"`
	Highest float64 `json:"H"`
	Lowest  float64 `json:"L"`
	Volume  float64 `json:"V"`
	BVolume float64 `json:"BV"`
	Time    string  `json:"T"`
}

type Response struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Result  []Candle `json:"result"`
}

type Listing struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}
