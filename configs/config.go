package configs

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	LNDHost          string
	MacaroonLocation string
	Macaroon         string

	DeezyPeer              string
	LoopSizeMinSat         string
	LocalAmountMinSat      string
	MaxLiqFeePpm           string
	ExcludeDeezyFromLiqOps string
	PayTimeoutSeconds      string

	KrakenEnabled        string
	KrakenOpMinBtc       string
	KrakenOpMaxBtc       string
	KrakenWithdrawBtcMin string
	KrakenUsername       string
	KrakenPassword       string
	KrakenOtpRequired    string
	KrakenOtpSecret      string
	KrakenApiKey         string
	KrakenApiSecret      string
	ChromeProfilePath    string

	StrikeEnabled                   string
	StrikeRepurchaserManualMode     string
	StrikeOpMinBtc                  string
	StrikeOpMaxBtc                  string
	StrikeWithdrawBtcMin            string
	StrikeDailyLimitBufferUsd       string
	StrikeWeeklyLimitBufferUsd      string
	StrikeRepurchaseCooldownSeconds string
	StrikeApiKey                    string
	StrikeJwtToken                  string
}

func GetConfig(ctx context.Context) (configs Config) {
	return ctx.Value("configs").(Config)
}

func LoadConfig(ctx context.Context) (context.Context, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		return ctx, err
	}

	configs := Config{
		LNDHost:          os.Getenv("LND_HOST"),
		MacaroonLocation: os.Getenv("MACAROON_LOCATION"),
		Macaroon:         os.Getenv("MACAROON"),

		DeezyPeer:              os.Getenv("DEEZY_PEER"),
		LoopSizeMinSat:         os.Getenv("LOOP_SIZE_MIN_SAT"),
		LocalAmountMinSat:      os.Getenv("LOCAL_AMOUNT_MIN_SAT"),
		MaxLiqFeePpm:           os.Getenv("MAX_LIQ_FEE_PPM"),
		ExcludeDeezyFromLiqOps: os.Getenv("EXCLUDE_DEEZY_FROM_LIQ_OPS"),
		PayTimeoutSeconds:      os.Getenv("PAY_TIMEOUT_SECONDS"),

		KrakenEnabled:        os.Getenv("KRAKEN_ENABLED"),
		KrakenOpMinBtc:       os.Getenv("KRAKEN_OP_MIN_BTC"),
		KrakenOpMaxBtc:       os.Getenv("KRAKEN_OP_MAX_BTC"),
		KrakenWithdrawBtcMin: os.Getenv("KRAKEN_WITHDRAW_BTC_MIN"),
		KrakenUsername:       os.Getenv("KRAKEN_USERNAME"),
		KrakenPassword:       os.Getenv("KRAKEN_PASSWORD"),
		KrakenOtpRequired:    os.Getenv("KRAKEN_OTP_REQUIRED"),
		KrakenOtpSecret:      os.Getenv("KRAKEN_OTP_SECRET"),
		KrakenApiKey:         os.Getenv("KRAKEN_API_KEY"),
		KrakenApiSecret:      os.Getenv("KRAKEN_API_SECRET"),
		ChromeProfilePath:    os.Getenv("CHROME_PROFILE_PATH"),

		StrikeEnabled:                   os.Getenv("STRIKE_ENABLED"),
		StrikeRepurchaserManualMode:     os.Getenv("STRIKE_REPURCHASER_MANUAL_MODE"),
		StrikeOpMinBtc:                  os.Getenv("STRIKE_OP_MIN_BTC"),
		StrikeOpMaxBtc:                  os.Getenv("STRIKE_OP_MAX_BTC"),
		StrikeWithdrawBtcMin:            os.Getenv("STRIKE_WITHDRAW_BTC_MIN"),
		StrikeDailyLimitBufferUsd:       os.Getenv("STRIKE_DAILY_LIMIT_BUFFER_USD"),
		StrikeWeeklyLimitBufferUsd:      os.Getenv("STRIKE_WEEKLY_LIMIT_BUFFER_USD"),
		StrikeRepurchaseCooldownSeconds: os.Getenv("STRIKE_REPURCHASER_COOLDOWN_SECONDS"),
		StrikeApiKey:                    os.Getenv("STRIKE_API_KEY"),
		StrikeJwtToken:                  os.Getenv("STRIKE_JWT_TOKEN"),
	}

	ctx = context.WithValue(ctx, "configs", configs)

	return ctx, nil
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
