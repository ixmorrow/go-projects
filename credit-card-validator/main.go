package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CardInfo struct {
	CardNumber string `json:"cardNumber"`
}

func luhnAlgorithm(input string) bool {
	// Convert the input string to a slice of integers
	digits := make([]int, len(input))

	for i, char := range input {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			// Return false if the input contains non-numeric characters
			return false
		}
		digits[i] = digit
	}

	// Double every second digit from the right and subtract 9 if the result is greater than 9
	for i := len(digits) - 2; i >= 0; i -= 2 {
		doubled := digits[i] * 2
		if doubled > 9 {
			doubled -= 9
		}
		digits[i] = doubled
	}

	// calculate the sum of all digits
	sum := 0
	for _, digit := range digits {
		sum += digit
	}

	// Check if the sum is a multiple of 10
	return sum%10 == 0
}

func validateCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var cardInfo CardInfo
	_ = json.NewDecoder(r.Body).Decode(&cardInfo)
	fmt.Println("Card number received:", cardInfo.CardNumber)
	isValidCardNumber := luhnAlgorithm(cardInfo.CardNumber)
	json.NewEncoder(w).Encode(isValidCardNumber)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/validateCreditCard", validateCard).Methods("GET")

	fmt.Println("Starting server at port 8000...")
	log.Fatal(http.ListenAndServe(":8000", r))
}
