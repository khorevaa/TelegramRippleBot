package main

import (
	"time"

	"telegram-bot-api"
)

type Configuration struct {
	BotToken, MetricToken                                                              string
	ChannelId, ChatId                                                                  int64
	BuySellXRP                                                                         string
	TwitterAccounts                                                                    []string
	TwitterAccessToken, TwitterAccessSecret, TwitterConsumerKey, TwitterConsumerSecret string
	RippleUrlBase                                                                      string
	RippleUrlParams                                                                    string
	BittrexChartURL                                                                    string
	CoinMarketCapListings                                                              string
	AdminIds                                                                           []int64
	ChannelHours, GroupHours, UsersHours, TwitterHours                                 int
	TwitterShareURL, SupportURL                                                        string
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
	TransactionResult string
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
	PhotoIds map[string]string
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

type Prices struct {
	LastCheck string  `json:"LastCheck"`
	Highs     Periods `json:"Highs"`
	Lows      Periods `json:"Lows"`
}

type Periods struct {
	AllTime     float64   `json:"allTime"`
	Week        []float64 `json:"7d"`
	Month       []float64 `json:"30d"`
	ThreeMonths []float64 `json:"3m"`
}

type PendingPost struct {
	Message  tgbotapi.Message
	PostTime time.Time
	//IsRepeating bool
	DelayHours  float64
	Destination int64
}

type TimeForSending struct {
	GroupTime, ChannelTime, UsersTime, TwitterTime time.Time
}

func (t TimeForSending) anyTime() bool {
	if time.Now().After(t.GroupTime) ||
		time.Now().After(t.ChannelTime) ||
		time.Now().After(t.UsersTime) ||
		time.Now().After(t.TwitterTime) {
		return true
	}
	return false
}
