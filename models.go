package main

import "time"

type Config struct {
	BotToken string
	ChannelId int64
	TwitterAccounts  []string
	TwitterAccessToken, TwitterAccessSecret, TwitterConsumerKey, TwitterConsumerSecret string
	RippleUrlBase string
	RippleUrlParams string
	RippleStatsUrl string
	BittrexChartURL string
}

type User struct {
	ID        int64      `gorm:"primary_key"`
	FirstName string
	LastName  string
	UserName  string
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

type Transaction struct{
	Date string `json:"date"`
	Hash string `json:"hash"`
	Tx TxInfo `json:"tx"`

}

type TxInfo struct{
	TransactionType, Destination string
	Amount string
}

//type AmountInfo struct {
//	Value, Currency, Issuer string
//}

type CachedStats struct {
	Time time.Time
	Stats string
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
