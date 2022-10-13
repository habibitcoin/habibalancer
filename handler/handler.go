package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/habibitcoin/habibalancer/configs"
	"github.com/habibitcoin/habibalancer/lightning"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

type (
	Handler struct {
		Context context.Context
		Config  configs.Config
	}
)

func (h *Handler) Index(c echo.Context) (err error) {
	// refresh context and configs
	h.Context = context.WithValue(h.Context, "configs", h.Config)

	var (
		configMap       = configs.GetConfig(h.Context).GetConfigMap()
		lightningClient = lightning.NewClient(h.Context)
		depositAddress  = "Invalid LND URL and/or macaroon"
		totalBalance    = "Invalid LND URL and/or macaroon"
	)

	depositAddressValid, err := lightningClient.CreateAddress()
	if err == nil {
		depositAddress = depositAddressValid
	}

	balance, err := lightningClient.GetBalance()
	if err == nil {
		totalBalance = balance.TotalBalance
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"configMap": configMap,
		"address":   depositAddress,
		"balance":   totalBalance,
	})
}

func (h *Handler) SaveConfig(c echo.Context) (err error) {
	newConfig := new(configs.Config)

	if err = c.Bind(newConfig); err != nil {
		log.Println(err)
		return nil
	}

	h.Config = *newConfig
	log.Println(h.Config.LNDHost)

	var (
		configMap       = newConfig.GetConfigMap()
		lightningClient = lightning.NewClient(h.Context)
		depositAddress  = "Invalid LND URL and/or macaroon"
		totalBalance    = "Invalid LND URL and/or macaroon"
	)

	// save/update .env file
	err = godotenv.Write(configMap, ".env")
	if err != nil {
		log.Printf("Error saving .env file: %v", err)
		return err
	}

	depositAddressValid, err := lightningClient.CreateAddress()
	if err == nil {
		depositAddress = depositAddressValid
	}

	balance, err := lightningClient.GetBalance()
	if err == nil {
		totalBalance = balance.TotalBalance
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"configMap": configMap,
		"address":   depositAddress,
		"balance":   totalBalance,
	})
}
