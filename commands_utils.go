package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"strings"
)

func getRippleStats() string {
	resp, err := http.Get(configuration.RippleStatsUrl)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	price := json.Get(bodyBytes, 0, "price_usd").ToString()
	volume := json.Get(bodyBytes, 0, "24h_volume_usd").ToString()
	cap := json.Get(bodyBytes, 0, "market_cap_usd").ToString()

	return "Price: " + price + " USD\nVolume: " + volume + " USD\nCapitalization: " + cap + " USD"
}

func checkAddress(a string) bool {
	if string(a[0]) != "r" {
		return false
	}
	if len(a) > 35 || len(a) < 25 {
		return false
	}
	return true
}

func getCurrency(name string) string {
	for _, listing := range listings {
		if strings.ToLower(listing.Name) == strings.ToLower(name) ||
			strings.ToLower(listing.Symbol) == strings.ToLower(name) {
			return listing.Name
		}
	}
	return ""
}
