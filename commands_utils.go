package main

import (
	"net/http"
	"log"
	"io/ioutil"
)

func getRippleStats() string {
	resp, err := http.Get(configuration.RippleStatsUrl)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		log.Print(err)
	}

	price := json.Get(bodyBytes, 0, "price_usd").ToString()
	volume := json.Get(bodyBytes, 0, "24h_volume_usd").ToString()
	cap := json.Get(bodyBytes, 0, "market_cap_usd").ToString()

	return "Price: " + price + " USD\nVolume: " + volume + " USD\nCapitalization: " + cap + " USD"
}
