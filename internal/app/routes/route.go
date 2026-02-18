package routes

import (
	"github.com/gin-gonic/gin"
	userCtrl "github.com/vizucode/e-wallet-system/internal/app/usecase/user/controllers"
	walletCtrl "github.com/vizucode/e-wallet-system/internal/app/usecase/wallets/controllers"
)

func NewRoute(route *gin.Engine, userController userCtrl.UserController, walletController walletCtrl.WalletController) {
	route.GET("/users", userController.GetAllUsers)
	route.GET("/users/:user_id/wallets", walletController.GetUserWallets)
	route.POST("/wallets", walletController.CreateWallet)
	route.POST("/wallets/transfer", walletController.TransferWallet)
	route.POST("/wallets/:id/topup", walletController.TopUpWallet)
	route.POST("/wallets/:id/pay", walletController.PayWallet)
	route.POST("/wallets/:id/suspend", walletController.SuspendWallet)
	route.GET("/wallets/:id", walletController.GetWalletByID)
}
