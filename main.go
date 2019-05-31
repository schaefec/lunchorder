package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)


// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	ScrapeNow string `json:"scrapenow"`
}

// HelloPubSub consumes a Pub/Sub message.
func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	log.Println("web scrape has been triggered")
	return nil
}

type mockContext struct {
}

func (*mockContext) Deadline() (time.Time, bool) {
	return time.Now(), false
}

func (*mockContext) Done() <-chan struct{} {
	return make(chan struct{})
}

func (*mockContext) Err() error {
	return nil
}

func (*mockContext) Value(v interface{}) interface{} {
	return nil
}


func main() {
	//	HelloPubSub(&mockContext{}, PubSubMessage{
	//		ScrapeNow: "true",
	//	})

	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		log.Fatal(err)
	}

	proxyUrl, err := url.Parse("http://localhost:8080")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
	    },
		Jar: jar,
	}

	data := url.Values{}
	data.Set("Login_Name", "***********")
	data.Set("Login_Passwort", "***********")

	loginUrl, err := url.Parse("https://mensa3.johanniter-schulmensa.de/index.php?m=2;0&ear_a=akt_login")
	if err != nil {
		log.Fatal(err)
	}

	// login
	resp, err := client.PostForm(loginUrl.String(), data)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatal("Status was not good: ", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	weekSelectionForm := doc.Find("form[name=form_sel_datum]").First();
	selectableWeeks := weekSelectionForm.Find("select[name=sel_datum]").Find("option")
	if selectableWeeks.Size() == 0 {
		log.Fatal("Not detected any weeks to select")
	}

	var notAllDaysSelected bool

	selectableWeeks.Each(func(n int,
		selection *goquery.Selection) {
			if  week, exists := selection.Attr("value"); exists {

				target := "https://mensa3.johanniter-schulmensa.de/index.php?m=2;0"

				formData := url.Values{}
				formData.Set("sel_datum", week)

				req, err := http.NewRequest("POST", target, strings.NewReader(formData.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				resp, err := client.Do(req)
				if err != nil {
					log.Fatal("Unable to get week content: ", week, err)
				}
				defer resp.Body.Close()

				doc, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					log.Fatal(err)
				}

				//weekString, err := selection.Html()

				// Find all checkboxes that are not checked or all input fields of type text with value "0"
				if numOfNotSelected := doc.Find("form[name=speiseplan]").Find("input[type=checkbox]:not(" +
					"[checked=checked])," +
					"input[type=text][value=\"0\"]").Size(); numOfNotSelected > 0 {
					//fmt.Println(weekString)

					notAllDaysSelected = true
				}
			}
		})

		if notAllDaysSelected {
			fmt.Print("At least one week contains where not all days are selected.")
		}
}