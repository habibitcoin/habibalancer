package kraken

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	krakenapi "github.com/beldur/kraken-go-api-client"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/joho/godotenv"
)

var (
	api = krakenapi.New(GoDotEnvVariable("KRAKEN_API_KEY"), GoDotEnvVariable("KRAKEN_API_SECRET"))

	krakenWithdrawAmtXBTmin, _ = strconv.ParseFloat(GoDotEnvVariable("KRAKEN_WITHDRAW_BTC_MIN"), 64)
)

func GetBalance() (string, error) {
	result, err := api.Query("Balance", map[string]string{})
	if err != nil {
		log.Println("Unexpected error fetching Kraken balance: ", err)
		return "", err
	}
	res := result.(map[string]interface{})
	return fmt.Sprint(res["XXBT"]), nil
}

func Withdraw() (interface{}, error) {
	krakenBalanceStringXBT, err := GetBalance()
	if err != nil {
		log.Println("Error fetching Kraken balance")
		log.Println(err)
		return nil, err
	}
	log.Println("Kraken balance XBT")
	log.Println(krakenBalanceStringXBT)
	krakenBalanceFloatXBT, _ := strconv.ParseFloat(krakenBalanceStringXBT, 64)

	if krakenBalanceFloatXBT > krakenWithdrawAmtXBTmin {
		result, err := api.Query("Withdraw", map[string]string{
			"asset":  "xbt",
			"key":    "umbrel",
			"amount": krakenBalanceStringXBT,
		})
		if err != nil {
			log.Println("Unexpected error performing Kraken withdrawal")
			log.Println(err)
			return nil, err
		}
		return result, nil
	}
	log.Println("Balance too low for withdrawal " + krakenBalanceStringXBT)
	return nil, nil
}

// Receives an amount defined in BTC, returns an invoice
// NOTE: The first time you run this, you need.
func GetAddress(amount string) (invoice string) {
	// create chrome instance
	ctx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.WindowSize(25, 25),
			chromedp.Flag("headless", false), // Sorry, doesn't work headless
			chromedp.UserDataDir(GoDotEnvVariable("CHROME_PROFILE_PATH")))...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(
		ctx,
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	// Launch browser and visit
	var location string
	err := chromedp.Run(ctx,
		browser.SetPermission(&browser.PermissionDescriptor{Name: "clipboard-read"}, browser.PermissionSettingGranted).WithOrigin("https://www.kraken.com"),
		chromedp.Navigate(`https://www.kraken.com/u/funding/deposit?asset=BTC&method=1`), // may redirect us to login page
		chromedp.Sleep(3*time.Second),
		chromedp.Location(&location),
	)
	if err != nil {
		log.Println(err)
		return ""
	}

	if location != "https://www.kraken.com/u/funding/deposit?asset=BTC&method=1" {
		log.Println(location)
		// login is required
		err = chromedp.Run(ctx,
			// wait for footer element is visible (ie, page is loaded)
			chromedp.WaitVisible(`//input[@name="username"]`),
			chromedp.SendKeys(`//input[@name="username"]`, GoDotEnvVariable("KRAKEN_USERNAME")),
			chromedp.WaitVisible(`//input[@name="password"]`),
			chromedp.SendKeys(`//input[@name="password"]`, GoDotEnvVariable("KRAKEN_PASSWORD")),
			chromedp.Sleep(3*time.Second),
			chromedp.SendKeys(`//input[@name="password"]`, kb.Enter),
			// find and click body > reach-portal:nth-child(37) > div:nth-child(3) > div > div > div > div > div.tr.mt3 > button.Button_button__caA8R.Button_primary__c5lrD.Button_large__T4YrY.no-tab-highlight
			chromedp.Sleep(3*time.Second),
			chromedp.SendKeys(`//input[@name="tfa"]`, getHOTPToken(GoDotEnvVariable("KRAKEN_OTP_SECRET"))),
			chromedp.Sleep(1*time.Second),
			chromedp.SendKeys(`//input[@name="tfa"]`, kb.Enter),
			chromedp.Sleep(3*time.Second),
			chromedp.Navigate(`https://www.kraken.com/u/funding/deposit?asset=BTC&method=1`),
			chromedp.Sleep(3*time.Second),
			chromedp.Location(&location),
			// GO CONFIRM YOUR DEVICE VIA EMAIL, COMMENT THIS OUT AGAIN AND RESTART SCRIPT
		)
		if err != nil {
			log.Println(err)
			return ""
		}
	}

	if location != "https://www.kraken.com/u/funding/deposit?asset=BTC&method=1" {
		log.Println("You may need to confirm your email and restart!")
		os.Exit(1)
	}

	// navigate to a page, wait for an element, click
	err = chromedp.Run(ctx,
		chromedp.Sleep(10*time.Second),
		chromedp.Click(`div:nth-child(3) > div > div > div > div > div.tr.mt3 > button.Button_button__caA8R.Button_primary__c5lrD.Button_large__T4YrY.no-tab-highlight`, chromedp.ByQueryAll),
		chromedp.Sleep(2*time.Second),
		chromedp.SendKeys(`//input[@name="amount"]`, amount),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`#__next > div > main > div > div.container > div > div.FundingTransactionPage_form__OGaKV > div > div > div:nth-child(4) > div.LightningForm_callToAction__Y4b1E > button`, chromedp.NodeVisible),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`#__next > div > main > div > div.container > div > div.FundingTransactionPage_form__OGaKV > div > div > div:nth-child(4) > div:nth-child(5) > div > div > button > div`, chromedp.ByQuery),
		chromedp.Evaluate(`window.navigator.clipboard.readText()`, &invoice, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	)
	if err != nil {
		log.Println(err)
		return ""
	}
	return invoice
}

// use godot package to load/read the .env file and
// return the value of the key.
func GoDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func getHOTPToken(secret string) string {
	// Converts secret to base32 Encoding. Base32 encoding desires a 32-character
	// subset of the twenty-six letters A–Z and ten digits 0–9
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		log.Println(err)
		return ""
	}
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(time.Now().Unix()/30))

	// Signing the value using HMAC-SHA1 Algorithm
	hash := hmac.New(sha1.New, key)
	hash.Write(bs)
	h := hash.Sum(nil)

	// We're going to use a subset of the generated hash.
	// Using the last nibble (half-byte) to choose the index to start from.
	// This number is always appropriate as it's maximum decimal 15, the hash will
	// have the maximum index 19 (20 bytes of SHA1) and we need 4 bytes.
	o := (h[19] & 15)

	var header uint32
	// Get 32 bit chunk from hash starting at the o
	r := bytes.NewReader(h[o : o+4])
	err = binary.Read(r, binary.BigEndian, &header)

	if err != nil {
		log.Println(err)
		return ""
	}
	// Ignore most significant bits as per RFC 4226.
	// Takes division from one million to generate a remainder less than < 7 digits
	h12 := (int(header) & 0x7fffffff) % 1000000

	// Converts number as a string
	otp := strconv.Itoa(int(h12))

	return prefix0(otp)
}

func prefix0(otp string) string {
	if len(otp) == 6 {
		return otp
	}
	for i := 6 - len(otp); i > 0; i-- {
		otp = "0" + otp
	}
	return otp
}
