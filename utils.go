package main

import (
	"strconv"
	"log"
	"strings"
)

func intToString(i int) string {
	s := strconv.Itoa(i)
	return s
}

func stringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil{
		log.Print(err)
	}
	return i
}

func stringToFloat64(s string) float64 {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil{
		log.Print(err)
	}
	return i
}

func float64ToString(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	return s
}

func float64WithSign(f float64) string{
	if f >= 0{
		return "+"+float64ToString(f)
	}else {
		return float64ToString(f)
	}
}

func contains(words []string, word string) bool {
	for _, val := range words {
		if strings.Contains(word, val) {
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