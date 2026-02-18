package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vizucode/e-wallet-system/internal/app/dto/domains"
	"github.com/vizucode/e-wallet-system/internal/app/usecase/wallets/service"
)

type WalletController interface {
	GetUserWallets(c *gin.Context)
	CreateWallet(c *gin.Context)
	TopUpWallet(c *gin.Context)
	PayWallet(c *gin.Context)
	TransferWallet(c *gin.Context)
	SuspendWallet(c *gin.Context)
}

type walletController struct {
	walletService service.WalletService
}

func NewWalletController(walletService service.WalletService) WalletController {
	return &walletController{walletService: walletService}
}

func (ctrl *walletController) GetUserWallets(c *gin.Context) {
	userID := c.Param("user_id")

	result, err := ctrl.walletService.GetWalletsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve wallets",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *walletController) CreateWallet(c *gin.Context) {
	var req domains.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: please provide 'user_id' and 'currency' fields",
		})
		return
	}

	result, err := ctrl.walletService.CreateWallet(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserIDRequired),
			errors.Is(err, service.ErrCurrencyRequired),
			errors.Is(err, service.ErrInvalidCurrency):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrWalletExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an unexpected error occurred while creating the wallet, please try again later",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (ctrl *walletController) TopUpWallet(c *gin.Context) {
	walletID := c.Param("id")

	var req domains.TopUpWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: please provide 'amount' (as a string) and 'reference_id' fields",
		})
		return
	}

	result, err := ctrl.walletService.TopUpWallet(walletID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount),
			errors.Is(err, service.ErrAmountNotPositive),
			errors.Is(err, service.ErrAmountPrecision),
			errors.Is(err, service.ErrReferenceRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrWalletNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "wallet not found: please check the wallet ID and try again",
			})
		case errors.Is(err, service.ErrWalletSuspended):
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrDuplicateReference):
			c.JSON(http.StatusConflict, gin.H{
				"error": "this top-up has already been processed (duplicate reference_id)",
			})
		case errors.Is(err, service.ErrConcurrentUpdate):
			c.JSON(http.StatusConflict, gin.H{
				"error": "the wallet was updated by another request, please retry your top-up",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an unexpected error occurred while processing the top-up, please try again later",
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *walletController) PayWallet(c *gin.Context) {
	walletID := c.Param("id")

	var req domains.PayWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: please provide 'amount' (as a string) and 'reference_id' fields",
		})
		return
	}

	result, err := ctrl.walletService.PayWallet(walletID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount),
			errors.Is(err, service.ErrAmountNotPositive),
			errors.Is(err, service.ErrAmountPrecision),
			errors.Is(err, service.ErrReferenceRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrWalletNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "wallet not found: please check the wallet ID and try again",
			})
		case errors.Is(err, service.ErrWalletSuspended):
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrInsufficientBalance):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrDuplicateReference):
			c.JSON(http.StatusConflict, gin.H{
				"error": "this payment has already been processed (duplicate reference_id)",
			})
		case errors.Is(err, service.ErrConcurrentUpdate):
			c.JSON(http.StatusConflict, gin.H{
				"error": "the wallet was updated by another request, please retry your payment",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an unexpected error occurred while processing the payment, please try again later",
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *walletController) TransferWallet(c *gin.Context) {
	var req domains.TransferWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: please provide 'from_wallet_id', 'to_wallet_id', 'amount', and 'reference_id' fields",
		})
		return
	}

	result, err := ctrl.walletService.TransferWallet(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFromWalletRequired),
			errors.Is(err, service.ErrToWalletRequired),
			errors.Is(err, service.ErrSameWallet),
			errors.Is(err, service.ErrInvalidAmount),
			errors.Is(err, service.ErrAmountNotPositive),
			errors.Is(err, service.ErrAmountPrecision),
			errors.Is(err, service.ErrReferenceRequired),
			errors.Is(err, service.ErrCurrencyMismatch):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrWalletNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrWalletSuspended):
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrInsufficientBalance):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, service.ErrDuplicateReference):
			c.JSON(http.StatusConflict, gin.H{
				"error": "this transfer has already been processed (duplicate reference_id)",
			})
		case errors.Is(err, service.ErrConcurrentUpdate):
			c.JSON(http.StatusConflict, gin.H{
				"error": "a wallet was updated by another request, please retry your transfer",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an unexpected error occurred while processing the transfer, please try again later",
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *walletController) SuspendWallet(c *gin.Context) {
	walletID := c.Param("id")

	result, err := ctrl.walletService.SuspendWallet(walletID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWalletNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "wallet not found: please check the wallet ID and try again",
			})
		case errors.Is(err, service.ErrConcurrentUpdate):
			c.JSON(http.StatusConflict, gin.H{
				"error": "the wallet was updated by another request, please retry",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an unexpected error occurred while suspending the wallet, please try again later",
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}
