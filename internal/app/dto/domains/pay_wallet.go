package domains

type PayWalletRequest struct {
	Amount      string `json:"amount" binding:"required"`
	ReferenceID string `json:"reference_id" binding:"required"`
}

type PayWalletResponse struct {
	WalletID string `json:"wallet_id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
	Status   string `json:"status"`
}
