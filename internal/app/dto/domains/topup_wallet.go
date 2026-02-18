package domains

type TopUpWalletRequest struct {
	Amount      string `json:"amount" binding:"required"`
	ReferenceID string `json:"reference_id" binding:"required"`
}

type TopUpWalletResponse struct {
	WalletID string `json:"wallet_id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
	Status   string `json:"status"`
}
