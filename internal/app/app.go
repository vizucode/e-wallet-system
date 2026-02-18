package app

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vizucode/e-wallet-system/configs/database"
	"github.com/vizucode/e-wallet-system/internal/app/routes"
	userCtrl "github.com/vizucode/e-wallet-system/internal/app/usecase/user/controllers"
	userRepo "github.com/vizucode/e-wallet-system/internal/app/usecase/user/repository"
	userSvc "github.com/vizucode/e-wallet-system/internal/app/usecase/user/service"
	walletCtrl "github.com/vizucode/e-wallet-system/internal/app/usecase/wallets/controllers"
	walletRepo "github.com/vizucode/e-wallet-system/internal/app/usecase/wallets/repository"
	walletSvc "github.com/vizucode/e-wallet-system/internal/app/usecase/wallets/service"
)

func Run() {
	router := gin.Default()

	var (
		host     = os.Getenv("DB_HOST")
		port     = os.Getenv("DB_PORT")
		user     = os.Getenv("DB_USER")
		password = os.Getenv("DB_PASSWORD")
		dbname   = os.Getenv("DB_NAME")
	)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := database.NewDatabaseConnection(dsn)
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	userRepository := userRepo.NewUserRepository(db)
	userService := userSvc.NewUserService(userRepository)
	userController := userCtrl.NewUserController(userService)

	walletRepository := walletRepo.NewWalletRepository(db)
	walletService := walletSvc.NewWalletService(walletRepository)
	walletController := walletCtrl.NewWalletController(walletService)

	routes.NewRoute(router, userController, walletController)

	router.Run(":8080")
}
