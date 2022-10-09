package strike

type balanceType struct {
	Currency string     `json:"currency"`
	Balance  amountType `json:"prepaidBalance"`
}

type bucketType struct{}

type limitType struct {
	Currency       string     `json:"currency"`
	BTCDailyLimit  bucketType `json:"btcWithdrawalTotal1`
	BTCWeeklyLimit bucketType `json:"btcWithdrawalTotal2`
}

type infoResponse struct {
	Balances []balanceType `json:"balances"`
	Limits   []limitType   `json:"limits"`
}

func GetBalanceAndLimits() (string, error) {
	return "", nil
}
