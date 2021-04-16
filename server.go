package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

type Balance struct {
	TXIDS  	[]ApiData `json:"txids"`
	Conf  	float32  `json:"confirmed"`
	Uncon 	float32  `json:"unconfirmed"`
}
type ApiData struct {
	TXID          string `json:"txid"`
	Value         string `json:"value"`
	Confirmations int    `json:"confirmations"`
}
type ApiError struct {
	Error string
}
func main() {
	port := os.Getenv("PORT");
	router := mux.NewRouter();
	router.HandleFunc("/balance/{address}", Address);
	router.HandleFunc("/balance/", Address);
	http.ListenAndServe(":"+port, router);
}
func Address(w http.ResponseWriter, req *http.Request) {
	errMsg := []byte(`{ "error": "Address not found." }`);
	addrr := mux.Vars(req)["address"];
	w.Header().Set("Access-Control-Allow-Origin", "*");

	if addrr == "" {
		w.Write(errMsg);
		return;
	}

	url := "https://blockbook-bitcoin.tronwallet.me/api/v2/utxo/" + addrr;

	resp, err1 := http.Get(url);
	if err1 != nil {
		w.Write(errMsg);
		return;
	}

	defer resp.Body.Close();

	body, err := ioutil.ReadAll(resp.Body);

	if err != nil {
		w.Write(errMsg);
		return;
	}
	
	var Error ApiError;
	json.Unmarshal(body, &Error);

	if Error.Error != "" {
		w.Write(errMsg);
		return;
	}

	var fetchData []ApiData;
	json.Unmarshal(body, &fetchData);
	transactions, confirmed, unconfirmed := apiResult(fetchData);

	apiResult := Balance{
		TXIDS:  transactions,
		Conf:  confirmed,
		Uncon: unconfirmed,
	}

	result, err := json.Marshal(apiResult);
	if err != nil {
		w.Write(errMsg);
		return;
	}

	w.Write(result);
}
func apiResult(transactions []ApiData) ([]ApiData, float32, float32) {
	confirmed := 0.0
	unconfirmed := 0.0
	var txid []ApiData
	for index := range transactions {
		div := 100000000.0
		value, err := strconv.ParseFloat(transactions[index].Value, 32);
		value = value / div;
		transactions[index].Value = fmt.Sprintf("%e\n", value);
		txid = append(txid, transactions[index])
		
		if err != nil {
			panic("Error trying to parse string to int in apiResult().")
		}
		if transactions[index].Confirmations < 2 {
			unconfirmed += value
		} else {
			confirmed += value
		}
	}
	return txid, float32(confirmed), float32(unconfirmed)
}
