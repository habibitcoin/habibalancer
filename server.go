package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/habibitcoin/habibalancer/configs"
	"github.com/habibitcoin/habibalancer/deezy"
	"github.com/habibitcoin/habibalancer/handler"
	"github.com/habibitcoin/habibalancer/lightning"
	"github.com/habibitcoin/habibalancer/operators/kraken"
	"github.com/habibitcoin/habibalancer/operators/nicehash"
	"github.com/habibitcoin/habibalancer/operators/strike"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	var webServer string
	ctx := context.Background()
	ctx, err := configs.LoadConfig(ctx)
	if err != nil {
		log.Println("You need to create a .env file or use the web browser helper at localhost:1323")
		webServer = "true"
	} else {
		webServer = configs.GetConfig(ctx).WebServer
	}

	if webServer == "true" {
		// Echo instance
		e := echo.New()

		templates := make(map[string]*template.Template)
		templates["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/base.html"))
		e.Renderer = &TemplateRegistry{
			templates: templates,
		}

		// Middleware
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())

		// Initialize handler
		h := &handler.Handler{
			Context: ctx,
			Config:  configs.GetConfig(ctx),
		}

		// Route => handler
		e.GET("/", h.Index)
		e.POST("/", h.SaveConfig)
		e.GET("/begin", func(c echo.Context) error {
			// refresh context and configs
			ctx := context.Background()
			ctx, _ = configs.LoadConfig(ctx)
			go looper(ctx)
			return c.String(http.StatusOK, "Looping started! Monitor command line for errors.\n")
		})
		e.Static("/static", "static")
		e.File("/favicon.ico", "static/images/favicon.ico")

		// Start server
		e.Logger.Fatal(e.Start(":1323"))
	} else {
		looper(ctx)
	}
}

func looper(ctx context.Context) (err error) {
	var (
		config             = configs.GetConfig(ctx)
		deezyPeer          = config.DeezyPeer
		cooldownSeconds, _ = strconv.Atoi(config.LoopCooldownSeconds)
		strikeEnabled      = config.StrikeEnabled
		krakenEnabled      = config.KrakenEnabled
		nicehashEnabled    = config.NicehashEnabled

		minLoopSize, _    = strconv.Atoi(config.LoopSizeMinSat)
		localAmountMin, _ = strconv.Atoi(config.LocalAmountMinSat)

		krakenAmtXBTmin, _ = strconv.ParseFloat(config.KrakenOpMinBtc, 64)
		krakenAmtXBTmax, _ = strconv.ParseFloat(config.KrakenOpMaxBtc, 64)

		strikeAmtXBTmin, _ = strconv.ParseFloat(config.StrikeOpMinBtc, 64)
		strikeAmtXBTmax, _ = strconv.ParseFloat(config.StrikeOpMaxBtc, 64)

		nicehashAmtXBTmin, _ = strconv.ParseFloat(config.NicehashOpMinBtc, 64)
		nicehashAmtXBTmax, _ = strconv.ParseFloat(config.NicehashOpMaxBtc, 64)

		maxLiqFeePpm, _ = strconv.ParseFloat(config.MaxLiqFeePpm, 64)

		lightningClient = lightning.NewClient(ctx)
		krakenClient    = kraken.NewClient(ctx)
		strikeClient    = strike.NewClient(ctx)
		nicehashClient  = nicehash.NewClient(ctx)
	)
	deezy.CalculateEarnings(lightningClient)
	if strikeEnabled == "true" {
		address, err := lightningClient.CreateAddress()
		if err != nil {
			log.Println("Error from LND creating new address for Strike withdrawal")
			return err
		}
		go strike.StrikeRepurchaser(ctx, address)
	}
	firstRun := true
	for {
		if !firstRun {
			time.Sleep(time.Duration(cooldownSeconds) * time.Second)
		}
		firstRun = false
		// Step 1: Find if we have a channel opened with Deezy
		chanExists := lightningClient.IsChannelOpen(deezyPeer)
		log.Println(chanExists)

		// Step 2:  If we do not have an open channel, see if we have enough money to open one
		if !chanExists {
			Balance, err := lightningClient.GetBalance()
			if err != nil {
				log.Println("Unexpected error fetching on-chain balance")
				log.Println(err)
			}
			log.Println("Onchain balance")
			log.Println(Balance)

			// Step 3: Open Channel to Danny
			totalBalance, _ := strconv.Atoi(Balance.TotalBalance)
			if totalBalance > 16500000 {
				totalBalance = 16500000 // dont use wumbo channels
			}
			if totalBalance > minLoopSize {
				log.Println("Opening channel to Deezy")
				resp, err := lightningClient.CreateChannel(deezyPeer, totalBalance-500000) // leave 500000 cushion
				if err != nil {
					log.Println("Error opening channel")
					log.Println(err)
					continue
				}
				log.Println("Channel Opened Successfully!")
				log.Println(resp)
			}
		} else {
			// Check if our open channel with Deezy's local balance is less than minimum close satoshis
			channels, err := lightningClient.ListChannels(deezyPeer)
			if err != nil {
				log.Println("Unexpected error fetching channels")
				log.Println(err)
				continue
			}
			if len(channels.Channels) > 0 {
				// If our local balance is less than the minimum, lets get paid!
				balanceInt, _ := strconv.Atoi(channels.Channels[0].LocalBalance)
				if balanceInt < localAmountMin {
					result, err := deezy.CloseChannel(channels.Channels[0].ChannelPoint, lightningClient)
					if err != nil {
						log.Println("Error getting paid from Deezy")
						log.Println(err)
						continue
					}
					log.Println("Deezy paid us!")
					log.Println(result)
				}
			}
		}

		// Step 4: Attempt sequence of liquidity operations, starting with sends
		/* General logic should follow:
		a. Check how much local liquidity I have on each channel
		b. Based on the amount, start attempting liq operations via:
			i. Kraken
			ii. Strike
			iii. etc
		c. Check again if local liquidity is acceptable and if liq op balances exceed deezyAmt, and exceed deezyAmt per channel rate (1000ppm)
		d. Send funds back to ourselves
		*/

		// STAY IN LOOP UNTIL BALANCE OF OPERATORS IS > LIQUIDITY OPERATION AMOUNT
		if krakenEnabled == "true" {
			if krakenAmtXBTmax > 0 {
				// Fetch Kraken LN Deposit Address
				krakenAmtXBTi := krakenAmtXBTmin + rand.Float64()*(krakenAmtXBTmax-krakenAmtXBTmin)
				krakenAmtXBT := fmt.Sprintf("%.5f", krakenAmtXBTi)
				krakenAmtXBTfee := fmt.Sprintf("%.0f", krakenAmtXBTi*maxLiqFeePpm*100) // fee is in satoshis, we want at least 50% profit
				lnInvoice := krakenClient.GetAddress(krakenAmtXBT)
				if lnInvoice == "" {
					continue
				}
				log.Println(lnInvoice)
				// Try to pay invoice
				for consecutiveErrors := 0; consecutiveErrors <= 10; consecutiveErrors++ {
					_, err = lightningClient.SendPayReq(lnInvoice, krakenAmtXBTfee)
					if err != nil {
						log.Println(err)
						if consecutiveErrors == 9 {
							time.Sleep(900 * time.Second)
							continue
						}
					}
					consecutiveErrors = 11
				}
			}

			// Step 5: Withdraw funds from Kraken if we have enough money to begin a liq operation
			// Get our Kraken balance in XBT

			// Try to withdraw all Kraken BTC because operator balance > liq amount
			result, err := krakenClient.Withdraw()
			if err != nil {
				log.Println(err)
			} else {
				fmt.Printf("Kraken withdrawal successful: %+v\n", result)
			}
		}

		if nicehashEnabled == "true" {
			if nicehashAmtXBTmax > 0 {
				// Fetch Nicehash LN Deposit Address
				nicehashAmtXBTi := nicehashAmtXBTmin + rand.Float64()*(nicehashAmtXBTmax-nicehashAmtXBTmin)
				nicehashAmtXBT := fmt.Sprintf("%.5f", nicehashAmtXBTi)
				nicehashAmtXBTfee := fmt.Sprintf("%.0f", nicehashAmtXBTi*maxLiqFeePpm*100) // fee is in satoshis, we want at least 50% profit
				lnInvoice, err := nicehashClient.GetAddress(nicehashAmtXBT)
				if lnInvoice == "" || err != nil {
					continue
				}
				log.Println(lnInvoice)
				// Try to pay invoice
				for consecutiveErrors := 0; consecutiveErrors <= 10; consecutiveErrors++ {
					_, err = lightningClient.SendPayReq(lnInvoice, nicehashAmtXBTfee)
					if err != nil {
						log.Println(err)
						if consecutiveErrors == 9 {
							time.Sleep(900 * time.Second)
							continue
						}
					}
					consecutiveErrors = 11
				}
			}
			// Step 5: Withdraw funds from Nicehash if we have enough money to begin a liq operation
			// Get our Nicehash balance in XBT

			// Try to withdraw all Nicehash BTC because operator balance > liq amount
			result, err := nicehashClient.Withdraw()
			if err != nil {
				log.Println(err)
			} else {
				fmt.Printf("Nicehash withdrawal successful: %+v\n", result)
			}
		}

		if strikeEnabled == "true" {
			// Begin Strike Liquidity Operation attempt
			if strikeAmtXBTmax > 0 {
				strikeAmtXBTi := strikeAmtXBTmin + rand.Float64()*(strikeAmtXBTmax-strikeAmtXBTmin)
				strikeAmtXBT := fmt.Sprintf("%.5f", strikeAmtXBTi)
				strikeAmtXBTfee := fmt.Sprintf("%.0f", strikeAmtXBTi*maxLiqFeePpm*100) // fee is in satoshis, we want at least 50% profit
				lnInvoice := strikeClient.GetAddress(strikeAmtXBT)
				if lnInvoice == "" {
					continue
				}
				log.Println(lnInvoice)
				// Try to pay invoice
				for consecutiveErrors := 0; consecutiveErrors <= 10; consecutiveErrors++ {
					_, err = lightningClient.SendPayReq(lnInvoice, strikeAmtXBTfee)
					if err != nil {
						log.Println(err)
						if consecutiveErrors == 9 {
							time.Sleep(900 * time.Second)
							continue
						}
					}
					consecutiveErrors = 11
				}
			}

			// Withdraw funds from Strike if we have a BTC account. USDT account do not have BTC accounts currently
			if strikeClient.DefaultCurrency == "USD" {
				success, err := strikeClient.Withdraw(lightningClient)
				if err != nil || success == false {
					log.Println(err)
					continue
				}
				log.Println("Strike withdrawal successful")
			}
		}
		time.Sleep(15 * time.Second)
	}

	return nil
}

// Define the template registry struct
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}
