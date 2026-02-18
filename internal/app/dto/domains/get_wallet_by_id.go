package domains

type GetWalletByIDResponse struct {
	WalletID string `json:"wallet_id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
	Status   string `json:"status"`
}
