package main

import (
	"strconv"
	"log"
	"os"
	"io/ioutil"
	"bytes"
	"github.com/m90/go-chatbase"
)

func int64ToString(i int64) string {
	s := strconv.FormatInt(i, 10)
	return s
}

func stringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Print(err)
	}
	return i
}

func stringToFloat64(s string) float64 {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Print(err)
	}
	return i
}

func float64ToString(f float64) string {
	f2 := float64(int(f*100))/100
	s := strconv.FormatFloat(f2, 'f', -1, 64)
	return s
}

func float64ToStringPrec3(f float64) string {
	s := strconv.FormatFloat(f, 'f', 3, 64)
	return s
}


func float64WithSign(f float64) string {
	if f >= 0 {
		return "+" + float64ToString(f)
	} else {
		return float64ToString(f)
	}
}

func contains(words []string, word string) bool {
	for _, val := range words {
		if word == val {
			return true
		}
	}
	return false
}

func containsInt64(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func readJson(obj interface{}, filename string) {
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		log.Print(err)
	}
	body, err := ioutil.ReadAll(file)
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	reader := bytes.NewReader(body)
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(obj)
	if err != nil {
		log.Print(err)
	}
	file.Close()
}

func writeJson(obj interface{}, filename string) {
	dataJson, err := json.Marshal(obj)
	if err != nil {
		log.Print(err)
	}
	ioutil.WriteFile(filename, dataJson, 0644)
}

func sendMetric(userId int, intent, message string) {
	botMessage := metric.Message(chatbase.UserType,
		strconv.Itoa(userId), chatbase.PlatformTelegram).
		SetIntent(intent).
		SetMessage(message)
	response, err := botMessage.Submit()
	if err != nil || !response.Status.OK() {
		log.Print(err)
		log.Print(response)
	}
}