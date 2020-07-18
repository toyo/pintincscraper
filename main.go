package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/sclevine/agouti"
)

func main() {

	var (
		id           = flag.String(`id`, ``, `メールアドレス`)
		password     = flag.String(`password`, ``, `パスワード`)
		startdatestr = flag.String(`startdate`, ``, `出力開始日`)
		enddatestr   = flag.String(`enddate`, ``, `出力終了日`)
	)

	flag.Parse()

	startdate, err := time.Parse(`20060102`, *startdatestr)
	switch {
	case err != nil:
		fmt.Println(err)
		startdate = time.Now()
	}

	enddate, err := time.Parse(`20060102`, *enddatestr)
	switch {
	case err != nil:
		fmt.Println(err)
		enddate = time.Now()
	}

	driver := agouti.ChromeDriver()
	if err := driver.Start(); err != nil {
		log.Fatalf("ブラウザ(Selenium Webdriver)が見つかりません: %v", err)
	}
	defer driver.Stop()

	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		log.Fatalf("ブラウザが開けません: %v", err)
	}

	if err := page.Navigate("https://pintinc.jp/login"); err != nil {
		log.Fatalf("ページが表示できません: %v", err)
	}

	err = page.FindByName(`email`).SendKeys(*id)
	if err != nil {
		log.Fatal(err)
	}
	err = page.FindByName(`password`).SendKeys(*password)
	if err != nil {
		log.Fatal(err)
	}
	err = page.FindByButton("ログイン").Submit()
	if err != nil {
		log.Fatal(err)
	}

	if err := page.Navigate("https://pintinc.jp/mypage/electricity/"); err != nil {
		log.Fatalf("ページが表示できません: %v", err)
	}

	err = page.FindByButton(`日別・時間別使用量`).Click()
	if err != nil {
		log.Fatal(err)
	}
	err = page.FindByLink(`時間別`).Click()
	if err != nil {
		log.Fatal(err)
	}

	for {

		var str string
		str, err = page.FindByName(`yyyymmdd`).Attribute(`value`)
		if err != nil {
			log.Fatal(err)
		}

		var dt time.Time
		dt, err = time.Parse(`20060102`, str)
		if err != nil {
			log.Fatal(err)
		}

		if !dt.Before(startdate) && !dt.After(enddate) {

			fmt.Print(dt.Format(`2006/01/02`))

			amount := make([]float64, 24)

			var val []struct {
				Label   string          `json:"label"`
				Amounts [][]interface{} `json:"amounts"`
			}
			{
				container := page.FindByID(`container`)
				if n, err := container.Count(); n == 0 {
					fmt.Print(`,NODATA`)
					break // nodata
				} else if err != nil {
					log.Fatal(err)
				}
				var str string
				str, err = container.Attribute(`data-consumption-mass`)
				if err != nil {
					log.Fatal(err)
				}
				err = json.Unmarshal([]byte(str), &val)
				if err != nil {
					log.Fatal(err)
				}
			}
			for _, v := range val { //  昼間 / 夜間
				for _, vv := range v.Amounts { // per hour
					if am, ok := vv[1].(float64); ok {
						if tm, err := strconv.Atoi(strings.TrimRight(vv[0].(string), `時`)); err != nil {
							log.Fatal(err)
						} else {
							amount[tm] = am
						}
					}
				}
			}

			for _, v := range amount {
				fmt.Print(`,`)
				fmt.Print(v)
			}
			fmt.Println()
		}
		err = page.FindByClass(`arrow-left`).Click()
		if err != nil {
			log.Fatal(err)
		}
	}
	return

}
