package domains

type SuspendWalletResponse struct {
	WalletID string `json:"wallet_id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
	Status   string `json:"status"`
}
