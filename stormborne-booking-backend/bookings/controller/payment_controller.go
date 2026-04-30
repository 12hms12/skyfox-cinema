package controller

import (
	"fmt"
	"net/http"
	"skyfox/bookings/constants"
	"skyfox/bookings/dto/request"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func HandlePayment(ctx *gin.Context) {
	var request request.PaymentGatewayRequest

	err := ctx.Bind(&request)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "check the request and try again"})
		return
	}

	if len(request.CardNumber) != constants.CARD_LENGTH {
		errMsg := "Payment Failed: Card Number should be 16 digits long!"
		ctx.JSON(400, gin.H{
			"message": errMsg,
		})
		return
	}

	cvv := strconv.Itoa(request.CVV)
	if len(cvv) != constants.CVV_LENGTH {
		errMsg := "Payment Failed: cvv should be 3 digit long!"
		ctx.JSON(400, gin.H{
			"message": errMsg,
		})
		return
	}

	name := request.Name
	if len(name) < constants.MIN_NAME_LENGTH {
		errMsg := "Payment Failed: Name on card should have min 3 characters!"
		ctx.JSON(400, gin.H{
			"message": errMsg,
		})
		return
	}

	expiryDate := request.ExpiryDate
	if !isWithinExpiryDate(expiryDate) {
		errMsg := "Payment Failed: Card Expired!"
		ctx.JSON(400, gin.H{
			"message": errMsg,
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Payment Successful",
	})
}

func isWithinExpiryDate(expiryDate string) bool {

	currentDate := time.Now()
	currentYear := currentDate.Year() % 100
	currentMonth := int(currentDate.Month())

	parts := strings.Split(expiryDate, "/")
	expiryMonth, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Println("Error parsing expiry month:", err)
		return false
	}
	expiryYear, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Println("Error parsing expiry year:", err)
		return false
	}

	return expiryYear >= currentYear && expiryMonth >= 1 && expiryMonth <= 12 && !(expiryYear == currentYear && expiryMonth < currentMonth)
}
