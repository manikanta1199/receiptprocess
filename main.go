package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Item struct {
	Description string `json:"shortDescription"`
	Price       string `json:"price"`
}

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

var ReceiptMap map[string]Receipt = make(map[string]Receipt)

// POST Handler
func CreateReceiptHandler(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := uuid.New().String()
	ReceiptMap[id] = receipt
	data := make(map[string]string)
	data["id"] = id
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetReceiptPointsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	receipt := ReceiptMap[id]
	points := 0
	for _, r := range receipt.Retailer {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			points++
		}
	}
	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if math.Mod(total*100, 10.00) == 0 {
		points += 50
	}
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}
	points += int(len(receipt.Items)/2) * 5
	for _, v := range receipt.Items {
		if len(strings.Trim(v.Description, " "))%3 == 0 {
			price, _ := strconv.ParseFloat(v.Price, 64)
			points += int(math.Ceil(price * 0.2))
		}
	}
	pDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
	if pDate.Day()%2 != 0 {
		points += 6
	}
	ptime, _ := time.Parse("15:04", receipt.PurchaseTime)
	if ptime.Hour() >= 14 && ptime.Hour() < 16 && ptime.Minute() > 0 {

		points += 10
	}
	data := make(map[string]int)
	data["points"] = points
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/receipt/{id}/points", GetReceiptPointsHandler).Methods("GET")
	r.HandleFunc("/receipts/process", CreateReceiptHandler).Methods("POST")
	fmt.Println("Server starting on port 4000...")
	if err := http.ListenAndServe(":4000", r); err != nil {
		log.Fatal(err)
	}
}
