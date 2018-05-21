package main

import (
	"github.com/fogleman/gg"
	"image/color"
	"net/http"
	"log"
	"time"
	"strconv"
	"fmt"
	"strings"
)

var (
	dc                     *gg.Context
	width                  = float64(1000)
	height                 = float64(375)
	chartHeight            = float64(336)
	volumesHeight          = float64(80)
	candleWidth            = float64(12)
	candleAreaWidth        = float64(18)
	hValue, lValue, vValue float64
	backgroundColor        = color.RGBA{224, 217, 201, 255}
	whiteColor             = color.RGBA{255, 255, 255, 255}
	strokeColor            = color.RGBA{192, 186, 172, 255}
	blackColor             = color.RGBA{0, 0, 0, 200}
	fontColor              = color.RGBA{77, 77, 77, 255}
	blackColorTrans        = color.RGBA{0, 0, 0, 75}
	greenColor             = color.RGBA{74, 193, 113, 255}
	greenColorTrans        = color.RGBA{141, 216, 166, 255}
	redColor               = color.RGBA{241, 37, 43, 255}
	redColorTrans          = color.RGBA{248, 133, 136, 255}
	candles                []Candle
	font                   = "HelveticaNeueBold.ttf"
)

func loadChart(period string) {
	pair := "usdt-xrp"
	candles = getCandlesBittrex(pair, period)
	dc = gg.NewContext(int(width), int(height))
	drawBackground()
	//drawLogo()
	calculateBounds(candles)
	drawValues()
	drawFrame()
	for i, val := range candles {
		drawCandle(float64(i), val, period)
	}
	drawCurrentValue()
	drawName(pair)
	drawDate()
	dc.SavePNG("chart-" + period + ".png")
}

func drawBackground() {
	dc.SetColor(backgroundColor)
	dc.Clear()
}

func drawLogo() {
	im, err := gg.LoadPNG("logo.png")
	if err != nil {
		log.Print(err)
	}
	dc.DrawImage(im, 0, 0)
}

func drawValues() {
	diff := hValue - lValue
	delta := diff / 5
	deltaPx := float64(7)
	current := hValue
	for i := 0; i <= 5; i++ {
		dc.SetColor(fontColor)
		dc.LoadFontFace(font, 13)
		var prec int
		if candles[len(candles)-1].Close > 1 {
			prec = 3
		} else {
			prec = 8
		}
		dc.DrawStringAnchored(strconv.FormatFloat(current, 'f', prec, 64), 916, deltaPx, 0, 0.5)
		dc.DrawLine(909, deltaPx, 913, deltaPx)
		dc.SetColor(blackColor)
		dc.SetLineWidth(1)
		dc.Stroke()
		dc.DrawLine(0, deltaPx, 907, deltaPx)
		dc.SetColor(strokeColor)
		dc.SetLineWidth(1)
		dc.Stroke()
		deltaPx += float64(chartHeight / 5)
		current -= delta
	}
}

func calculateBounds(candles []Candle) {
	hValue = candles[0].Highest
	lValue = candles[0].Lowest
	vValue = candles[0].Volume
	for _, val := range candles {
		if val.Highest > hValue {
			hValue = val.Highest
		}
		if val.Lowest < lValue {
			lValue = val.Lowest
		}
		if val.Volume > vValue {
			vValue = val.Volume
		}
	}
}

func drawFrame() {
	dc.DrawLine(0, 350, 908, 350)
	dc.SetColor(blackColor)
	dc.SetLineWidth(1)
	dc.Stroke()
	dc.DrawLine(908, 0, 908, 350)
	dc.SetColor(blackColor)
	dc.SetLineWidth(1)
	dc.Stroke()
}

func drawCandle(i float64, candle Candle, period string) {
	open := candle.Open
	close := candle.Close
	highest := candle.Highest
	lowest := candle.Lowest
	volume := candle.Volume

	//время и сетка
	t1, err := time.Parse(
		"2006-01-02T15:04:05",
		candle.Time)
	if err != nil {
		val := stringToInt64(candle.Time)
		t1 = time.Unix(int64(val)/1000, 0).UTC()
	}

	if int(i+1)%5 == 0 {
		dc.SetColor(fontColor)
		if err = dc.LoadFontFace(font, 13); err != nil {
			log.Print(err)
		}

		if period == "day" {
			day := int64ToString(int64(t1.Day()))
			month := int64ToString(int64(t1.Month()))

			dc.DrawStringAnchored(day+"/"+month,
				7+i*candleAreaWidth+candleWidth/2, chartHeight+25, 0.5, 0.5)
		} else {
			hours := int64ToString(int64(t1.Hour()))
			var minutes string
			if t1.Minute() == 0 {
				minutes = "00"
			} else {
				minutes = int64ToString(int64(t1.Minute()))

			}
			dc.DrawStringAnchored(hours+":"+minutes,
				7+i*candleAreaWidth+candleWidth/2, chartHeight+25, 0.5, 0.5)
		}
		dc.DrawLine(7+i*candleAreaWidth+candleWidth/2,
			chartHeight+15, 7+i*candleAreaWidth+candleWidth/2, chartHeight+19)
		dc.SetColor(blackColor)
		dc.SetLineWidth(1)
		dc.Stroke()
		dc.DrawLine(7+i*candleAreaWidth+candleWidth/2,
			0, 7+i*candleAreaWidth+candleWidth/2, chartHeight+15)
		dc.SetColor(strokeColor)
		dc.SetLineWidth(1)
		dc.Stroke()
	}

	//объем
	dc.DrawRectangle(4+i*candleAreaWidth,
		chartHeight+14-(volume)*volumesHeight/(vValue), candleWidth+6,
		(volume)*volumesHeight/(vValue))
	if open >= close {
		dc.SetColor(redColorTrans)
	} else {
		dc.SetColor(greenColorTrans)
	}
	dc.FillPreserve()
	dc.SetColor(blackColorTrans)
	dc.SetLineWidth(1)
	dc.Stroke()

	//палка
	dc.DrawRectangle(7+i*candleAreaWidth+candleWidth/2,
		7+(hValue-highest)*chartHeight/(hValue-lValue), 1,
		(highest-lowest)*chartHeight/(hValue-lValue))
	dc.SetColor(blackColor)
	dc.Fill()

	//свеча
	dc.DrawRectangle(7+i*candleAreaWidth,
		7+(hValue-open)*chartHeight/(hValue-lValue), candleWidth,
		(open-close)*chartHeight/(hValue-lValue))
	if open >= close {
		dc.SetColor(redColor)
	} else {
		dc.SetColor(greenColor)
	}
	dc.FillPreserve()
	dc.SetColor(blackColorTrans)
	dc.SetLineWidth(1)
	dc.Stroke()

}

func drawCurrentValue() {
	dc.DrawRectangle(910,
		(hValue-candles[len(candles)-1].Close)*chartHeight/(hValue-lValue), 90,
		16)
	if candles[len(candles)-1].Close > candles[len(candles)-1].Open {
		dc.SetColor(greenColor)
	} else {
		dc.SetColor(redColor)
	}

	dc.Fill()
	dc.MoveTo(910, (hValue-candles[len(candles)-1].Close)*chartHeight/(hValue-lValue))
	dc.LineTo(903, (hValue-candles[len(candles)-1].Close)*chartHeight/(hValue-lValue)+8)
	dc.LineTo(910, (hValue-candles[len(candles)-1].Close)*chartHeight/(hValue-lValue)+16)
	dc.ClosePath()
	dc.Fill()
	dc.SetColor(whiteColor)
	dc.LoadFontFace(font, 13)
	var prec int
	if candles[len(candles)-1].Close > 1 {
		prec = 3
	} else {
		prec = 8
	}
	dc.DrawStringAnchored(strconv.FormatFloat(candles[len(candles)-1].Close,
		'f', prec, 64), 916,
		(hValue-candles[len(candles)-1].Close)*chartHeight/(hValue-lValue)+7, 0, 0.5)

}

func drawName(request string) {
	dc.SetColor(blackColor)
	dc.DrawRectangle(5, 5, 90,
		25)
	dc.Fill()
	dc.SetColor(whiteColor)
	dc.DrawStringAnchored(strings.ToUpper(request), 50, 17, 0.5, 0.5)
}

func drawDate() {
	t1, err := time.Parse(
		"2006-01-02T15:04:05",
		candles[len(candles)-1].Time)
	if err != nil {
		val, _ := strconv.ParseUint(candles[len(candles)-1].Time, 10, 64)
		t1 = time.Unix(int64(val)/1000, 0).UTC()
	}
	dc.SetColor(blackColor)
	dc.DrawRectangle(100, 5, 130,
		25)
	dc.Fill()
	dc.SetColor(whiteColor)

	dc.DrawStringAnchored(fmt.Sprintf("%02d/%02d/%d UTC+0",
		t1.Month(), t1.Day(), t1.Year()), 165, 17, 0.5, 0.5)

}

func getCandlesBittrex(pair, period string) []Candle {
	var response Response
	var resp *http.Response
	var err error
	for {
		resp, err = http.Get(fmt.Sprintf(config.BittrexChartURL, period, pair))
		if err != nil {
			log.Print(err)
			continue
		}
		json.NewDecoder(resp.Body).Decode(&response)
		resp.Body.Close()
		if len(response.Result) != 0 {
			break
		}
	}
	if len(response.Result) <= 50 {
		return response.Result
	} else {
		return response.Result[len(response.Result)-50:]
	}
}
