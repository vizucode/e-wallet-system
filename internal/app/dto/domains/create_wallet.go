package domains

type CreateWalletRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

type CreateWalletResponse struct {
	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
	Status   string `json:"status"`
}
