package configs

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	LNDHost          string `form:"LNDHost"`
	MacaroonLocation string `form:"MacaroonLocation"`
	Macaroon         string `form:"Macaroon"`
	WebServer        string `form:"WebServer"`
	FeeRateSatsPerVb string `form:"FeeRateSatsPerVb"`

	DeezyPeer              string `form:"DeezyPeer"`
	DeezyClearnetHost      string `form:"DeezyClearnetHost"`
	DeezyTorHost           string `form:"DeezyTorHost"`
	LoopSizeMinSat         string `form:"LoopSizeMinSat"`
	LocalAmountMinSat      string `form:"LocalAmountMinSat"`
	MaxLiqFeePpm           string `form:"MaxLiqFeePpm"`
	ExcludeDeezyFromLiqOps string `form:"ExcludeDeezyFromLiqOps"`
	PayTimeoutSeconds      string `form:"PayTimeoutSeconds"`
	LoopCooldownSeconds    string `form:"LoopCooldownSeconds"`

	KrakenEnabled            string `form:"KrakenEnabled"`
	KrakenOpMinBtc           string `form:"KrakenOpMinBtc"`
	KrakenOpMaxBtc           string `form:"KrakenOpMaxBtc"`
	KrakenWithdrawBtcMin     string `form:"KrakenWithdrawBtcMin"`
	KrakenUsername           string `form:"KrakenUsername"`
	KrakenPassword           string `form:"KrakenPassword"`
	KrakenOtpRequired        string `form:"KrakenOtpRequired"`
	KrakenOtpSecret          string `form:"KrakenOtpSecret"`
	KrakenApiKey             string `form:"KrakenApiKey"`
	KrakenApiSecret          string `form:"KrakenApiSecret"`
	KrakenWithdrawAddressKey string `form:"KrakenWithdrawAddressKey"`
	ChromeProfilePath        string `form:"ChromeProfilePath"`

	StrikeEnabled                   string `form:"StrikeEnabled"`
	StrikeOpMinBtc                  string `form:"StrikeOpMinBtc"`
	StrikeOpMaxBtc                  string `form:"StrikeOpMaxBtc"`
	StrikeWithdrawBtcMin            string `form:"StrikeWithdrawBtcMin"`
	StrikeDailyLimitBufferUsd       string `form:"StrikeDailyLimitBufferUsd"`
	StrikeWeeklyLimitBufferUsd      string `form:"StrikeWeeklyLimitBufferUsd"`
	StrikeRepurchaseCooldownSeconds string `form:"StrikeRepurchaseCooldownSeconds"`
	StrikeApiKey                    string `form:"StrikeApiKey"`
	StrikeJwtToken                  string `form:"StrikeJwtToken"`
	StrikeDefaultCurrency           string `form:"StrikeDefaultCurrency"`

	NicehashEnabled            string `form:"NicehashEnabled"`
	NicehashOpMinBtc           string `form:"NicehashOpMinBtc"`
	NicehashOpMaxBtc           string `form:"NicehashOpMaxBtc"`
	NicehashWithdrawBtcMin     string `form:"NicehashWithdrawBtcMin"`
	NicehashApiKey             string `form:"NicehashApiKey"`
	NicehashApiSecret          string `form:"NicehashApiSecret"`
	NicehashWithdrawAddressKey string `form:"NicehashWithdrawAddressKey"`
	NicehashOrganizationId     string `form:"NicehashOrganizationId"`
}

func (configs Config) GetConfigMap() (configMap map[string]string) {
	inrec, _ := json.Marshal(configs)
	json.Unmarshal(inrec, &configMap)
	return configMap
}

func GetConfig(ctx context.Context) (configs Config) {
	return ctx.Value("configs").(Config)
}

func LoadConfig(ctx context.Context) (context.Context, error) {
	var err error
	err = godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file, falling back to .env.sample: %v", err)
		if fatalErr := godotenv.Load("env/.env.sample"); fatalErr != nil {
			// load file from bindata.go
			// create dependencies
			// data, _ := Asset("env/.env.sample")
			// os.WriteFile(".env.sample", data, 0644)

			files := AssetNames()

			for _, file := range files {
				log.Println(file)
				data, _ := Asset(file)

				dir, _ := filepath.Split(file)

				if _, err := os.Stat(dir); os.IsNotExist(err) {
					// your file does not exist
					os.MkdirAll(dir, 0700)
				}

				err := os.WriteFile(file, data, 0644)
				log.Println(err)
			}

			if fatalErr := godotenv.Load("env/.env.sample"); fatalErr != nil {
				log.Fatalf(fatalErr.Error())
			}
		}
	}

	configs := Config{
		LNDHost:          os.Getenv("LNDHost"),
		MacaroonLocation: os.Getenv("MacaroonLocation"),
		Macaroon:         os.Getenv("Macaroon"),
		WebServer:        os.Getenv("WebServer"),
		FeeRateSatsPerVb: os.Getenv("FeeRateSatsPerVb"),

		DeezyPeer:              os.Getenv("DeezyPeer"),
		DeezyClearnetHost:      os.Getenv("DeezyClearnetHost"),
		DeezyTorHost:           os.Getenv("DeezyTorHost"),
		LoopSizeMinSat:         os.Getenv("LoopSizeMinSat"),
		LocalAmountMinSat:      os.Getenv("LocalAmountMinSat"),
		MaxLiqFeePpm:           os.Getenv("MaxLiqFeePpm"),
		ExcludeDeezyFromLiqOps: os.Getenv("ExcludeDeezyFromLiqOps"),
		PayTimeoutSeconds:      os.Getenv("PayTimeoutSeconds"),
		LoopCooldownSeconds:    os.Getenv("LoopCooldownSeconds"),

		KrakenEnabled:            os.Getenv("KrakenEnabled"),
		KrakenOpMinBtc:           os.Getenv("KrakenOpMinBtc"),
		KrakenOpMaxBtc:           os.Getenv("KrakenOpMaxBtc"),
		KrakenWithdrawBtcMin:     os.Getenv("KrakenWithdrawBtcMin"),
		KrakenUsername:           os.Getenv("KrakenUsername"),
		KrakenPassword:           os.Getenv("KrakenPassword"),
		KrakenOtpRequired:        os.Getenv("KrakenOtpRequired"),
		KrakenOtpSecret:          os.Getenv("KrakenOtpSecret"),
		KrakenApiKey:             os.Getenv("KrakenApiKey"),
		KrakenApiSecret:          os.Getenv("KrakenApiSecret"),
		KrakenWithdrawAddressKey: os.Getenv("KrakenWithdrawAddressKey"),
		ChromeProfilePath:        os.Getenv("ChromeProfilePath"),

		StrikeEnabled:                   os.Getenv("StrikeEnabled"),
		StrikeOpMinBtc:                  os.Getenv("StrikeOpMinBtc"),
		StrikeOpMaxBtc:                  os.Getenv("StrikeOpMaxBtc"),
		StrikeWithdrawBtcMin:            os.Getenv("StrikeWithdrawBtcMin"),
		StrikeDailyLimitBufferUsd:       os.Getenv("StrikeDailyLimitBufferUsd"),
		StrikeWeeklyLimitBufferUsd:      os.Getenv("StrikeWeeklyLimitBufferUsd"),
		StrikeRepurchaseCooldownSeconds: os.Getenv("StrikeRepurchaseCooldownSeconds"),
		StrikeApiKey:                    os.Getenv("StrikeApiKey"),
		StrikeJwtToken:                  os.Getenv("StrikeJwtToken"),
		StrikeDefaultCurrency:           os.Getenv("StrikeDefaultCurrency"),

		NicehashEnabled:            os.Getenv("NicehashEnabled"),
		NicehashOpMinBtc:           os.Getenv("NicehashOpMinBtc"),
		NicehashOpMaxBtc:           os.Getenv("NicehashOpMaxBtc"),
		NicehashWithdrawBtcMin:     os.Getenv("NicehashWithdrawBtcMin"),
		NicehashApiKey:             os.Getenv("NicehashApiKey"),
		NicehashApiSecret:          os.Getenv("NicehashApiSecret"),
		NicehashWithdrawAddressKey: os.Getenv("NicehashWithdrawAddressKey"),
		NicehashOrganizationId:     os.Getenv("NicehashOrganizationId"),
	}

	ctx = context.WithValue(ctx, "configs", configs)

	return ctx, err
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
