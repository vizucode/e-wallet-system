package domains

type TransferWalletRequest struct {
	FromWalletID string `json:"from_wallet_id" binding:"required"`
	ToWalletID   string `json:"to_wallet_id" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
	ReferenceID  string `json:"reference_id" binding:"required"`
}

type TransferWalletDetail struct {
	WalletID string `json:"wallet_id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
	Status   string `json:"status"`
}

type TransferWalletResponse struct {
	FromWallet TransferWalletDetail `json:"from_wallet"`
	ToWallet   TransferWalletDetail `json:"to_wallet"`
}
