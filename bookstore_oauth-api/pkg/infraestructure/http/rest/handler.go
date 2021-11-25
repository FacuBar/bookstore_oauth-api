package rest

import (
	"net/http"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
	"github.com/gin-gonic/gin"
)

func Handler(ats ports.AcessTokenService) *gin.Engine {
	router := gin.Default()

	router.POST("/oauth/access_token", createAccessToken(ats))
	router.GET("/oauth/access_token/:access_token_id", getAccessToken(ats))

	return router
}

const (
	grantTypePassword = "password"
)

func createAccessToken(s ports.AcessTokenService) gin.HandlerFunc {
	type request struct {
		GrantType string `json:"grant_type"`

		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c *gin.Context) {
		var loginRequest request
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			restErr := rest_errors.NewBadRequestError("invalid request")
			c.JSON(restErr.Status(), restErr)
			return
		}

		if loginRequest.GrantType != grantTypePassword {
			restErr := rest_errors.NewBadRequestError("grant_type not supported")
			c.JSON(restErr.Status(), restErr)
			return
		}

		token, err := s.Create(loginRequest.Email, loginRequest.Password)
		if err != nil {
			c.JSON(err.Status(), err)
			return
		}
		c.JSON(http.StatusOK, token)
	}
}

func getAccessToken(s ports.AcessTokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenId := c.Param("access_token_id")

		at, err := s.GetById(tokenId)

		if err != nil {
			c.JSON(err.Status(), err)
			return
		}

		c.JSON(http.StatusOK, at)
	}
}
