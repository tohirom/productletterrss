package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/feeds"
)

type Letter struct {
	Total string `json:"total"`
	Data  []struct {
		ID            string `json:"id"`
		Letterdateint string `json:"letterdateint"`
		Lettercode    string `json:"lettercode"`
		Letterdate    string `json:"letterdate"`
		Lettercontent string `json:"lettercontent"`
		Lettermtm     string `json:"lettermtm"`
	} `json:"data"`
}

func makeRSS(url string) string {

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	//クエリパラメータ
	params := request.URL.Query()
	params.Add("limit", "10")
	request.URL.RawQuery = params.Encode()

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	//ar letters []Letter
	letters := new(Letter)

	if err := json.Unmarshal(body, letters); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
	}
	//fmt.Printf("Total: %v \n", letters.Total)

	now := time.Now()
	feed := &feeds.Feed{
		Title:       "Lenovo Productletter RSS",
		Link:        &feeds.Link{Href: "https://www.lenovojp.com/business/productletter/"},
		Description: "製品発表レターを参照・ダウンロードできます。",
		Author:      &feeds.Author{Name: "Lenovo", Email: "jmoiron@jmoiron.net"},
		Created:     now,
	}

	for _, letter := range letters.Data {
		//fmt.Printf("id: %v \n", letter.ID)
		t, e := time.Parse("2006-01-02", letter.Letterdate)

		if e != nil {
			fmt.Println(e)
		}

		item := &feeds.Item{
			Title:       letter.Lettercontent,
			Link:        &feeds.Link{Href: "http://jmoiron.net/blog/idiomatic-code-reuse-in-go/"},
			Description: letter.Lettercode,
			Created:     t,
		}
		feed.Items = append(feed.Items, item)
	}

	rss, err := feed.ToRss()

	if err != nil {
		log.Fatal(err)
	}

	return rss
}

func handleRequests(rss string) {
	http.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) {

		ua := r.Header.Get("User-Agent")
		fmt.Printf("user agent is: %s \n", ua)
		invocationid := r.Header.Get("X-Azure-Functions-InvocationId")
		fmt.Printf("invocationid is: %s \n", invocationid)
		fmt.Fprintf(w, rss)
	})
	httpInvokerPort, exists := os.LookupEnv("FUNCTIONS_HTTPWORKER_PORT")
	if exists {
		fmt.Println("FUNCTIONS_HTTPWORKER_PORT: " + httpInvokerPort)
	}
	log.Println("Go server Listening...on httpInvokerPort:", httpInvokerPort)
	log.Fatal(http.ListenAndServe(":"+httpInvokerPort, nil))
}

func main() {

	url := "https://www.lenovo-smb.com/productletter/letterlistjson.php?key1=&key2=&key3=&key4=&key5=&key6=&key7=&sdstring=TVRveU9qTTZORG8xT2pFd09qWTZOem80T2prPTo6TWpBeU1DMHdNeTB4Tmc9PTpNakF5TUMwd05pMHhOQT09OjpNQT09&_dc=1592061834856&page=1&start=0"

	rss := makeRSS(url)

	handleRequests(rss)
}
