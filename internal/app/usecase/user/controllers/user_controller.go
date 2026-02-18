package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vizucode/e-wallet-system/internal/app/usecase/user/service"
)

type UserController interface {
	GetAllUsers(c *gin.Context)
}

type userController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) UserController {
	return &userController{userService: userService}
}

func (ctrl *userController) GetAllUsers(c *gin.Context) {
	users, err := ctrl.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve users",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}
