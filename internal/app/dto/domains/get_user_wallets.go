package domains

type WalletResponse struct {
	WalletID string `json:"wallet_id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
	Status   string `json:"status"`
}

type GetUserWalletsResponse struct {
	UserID  string           `json:"user_id"`
	Wallets []WalletResponse `json:"wallets"`
}
