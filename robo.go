package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"robo/cryptodiniservice"
	"sort"
	"strings"
)

type UserID = int
type Coin = cryptodiniservice.Coin
type Portfolio = cryptodiniservice.Portfolio

type BuyOrders struct {
	Orders []Coin
}

type SellOrders struct {
	Orders []Coin
}

type Robo struct {
}

type cmcCoin struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type cmcCoins struct {
	Coin []cmcCoin `json:"data"`
}

type coin_USDT struct {
	Coin
	usdt float64
}

type AssetManagerService interface {
	Deposit(uid UserID, amount float64) BuyOrders
	Withdraw(uid UserID, amount float64) SellOrders
	GetPort(uid UserID) Portfolio
}

func (robo Robo) Deposit(uid UserID, amount float64) BuyOrders {
	// algorithm of this robo is buying 40%, 25%, 20%, 10%, 5% from top 5 coins in 24hr
	cmcCoins, _ := best5In24hr()
	coin1 := Coin{Symbol: cmcCoins.Coin[0].Symbol, Amount: amount * 0.4}
	coin2 := Coin{Symbol: cmcCoins.Coin[1].Symbol, Amount: amount * 0.25}
	coin3 := Coin{Symbol: cmcCoins.Coin[2].Symbol, Amount: amount * 0.2}
	coin4 := Coin{Symbol: cmcCoins.Coin[3].Symbol, Amount: amount * 0.1}
	coin5 := Coin{Symbol: cmcCoins.Coin[4].Symbol, Amount: amount * 0.05}
	buyOrders := BuyOrders{Orders: []Coin{coin1, coin2, coin3, coin4, coin5}}
	return buyOrders
}

func best5In24hr() (*cmcCoins, error) {
	request, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest", nil)
	request.Header.Set("X-CMC_PRO_API_KEY", "cafc2287-ecc7-46f9-9aeb-40509deb3f6b")

	q := request.URL.Query()
	q.Add("sort", "percent_change_24h")
	q.Add("limit", "5")
	request.URL.RawQuery = q.Encode()

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return nil, err
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		cmcCoins := cmcCoins{}
		json.Unmarshal(data, &cmcCoins)
		return &cmcCoins, nil
	}
}

func (robo Robo) Withdraw(uid UserID, amount float64) SellOrders {
	// algorithm for sell is to sell coins that have highest usdt value
	// in order to sell least type of coins possible
	port := cryptodiniservice.GetPort(uid)
	coins := port.Coins
	coinValuesMap, _ := getCoinsPriceInUSDT(coins)
	portWithValue := portWithUsdtValue(coins, coinValuesMap)
	sort.Slice(portWithValue, func(i, j int) bool {
		return portWithValue[i].usdt > portWithValue[j].usdt
	})

	neededAmount := amount
	var orders []Coin
	for i := 0; i < len(portWithValue); i++ {
		if neededAmount < portWithValue[i].usdt {
			usdtToWithdraw := portWithValue[i].Amount * neededAmount / portWithValue[i].usdt
			coin := Coin{Symbol: portWithValue[i].Symbol, Amount: usdtToWithdraw}
			orders = append(orders, coin)
			break
		} else {
			coin := Coin{Symbol: portWithValue[i].Symbol, Amount: portWithValue[i].Amount}
			orders = append(orders, coin)
			neededAmount -= portWithValue[i].usdt
		}
	}
	return SellOrders{Orders: orders}
}

func getCoinsPriceInUSDT(coins []Coin) (map[string]float64, error) {
	request, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest", nil)
	request.Header.Set("X-CMC_PRO_API_KEY", "cafc2287-ecc7-46f9-9aeb-40509deb3f6b")

	q := request.URL.Query()
	q.Add("symbol", coinsToSymbolsString(coins))
	q.Add("convert", "USDT")
	request.URL.RawQuery = q.Encode()

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return nil, err
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var coinsMap map[string]interface{}
		json.Unmarshal(data, &coinsMap)
		coinsMap = coinsMap["data"].(map[string]interface{})
		return coinsMapToCmcCoinUSDT(coinsMap), nil
	}
}

func coinsToSymbolsString(coins []Coin) string {
	var str strings.Builder
	str.WriteString(coins[0].Symbol)
	for i := 1; i < len(coins); i++ {
		str.WriteString(",")
		str.WriteString(coins[i].Symbol)
	}
	return str.String()
}

func coinsMapToCmcCoinUSDT(coinsMap map[string]interface{}) map[string]float64 {
	usdtMap := make(map[string]float64)
	for key := range coinsMap {
		value := coinsMap[key].(map[string]interface{})
		quotes := value["quote"].(map[string]interface{})
		usdt := quotes["USDT"].(map[string]interface{})
		price := usdt["price"].(float64)
		usdtMap[key] = price
	}
	return usdtMap
}

func portWithUsdtValue(coins []Coin, coinValuesMap map[string]float64) []coin_USDT {
	var coinWithUsdt []coin_USDT
	for _, coin := range coins {
		usdt := coinValuesMap[coin.Symbol] * coin.Amount
		coinWithUsdt = append(coinWithUsdt, coin_USDT{Coin: coin, usdt: usdt})
	}
	return coinWithUsdt
}

func (robo Robo) GetPort(uid UserID) Portfolio {
	// I don't think GetPort should be in robo
	// robo should get port from Cryptodini to calculate how to adjust port
	// not the other way around
	return Portfolio{}
}

func main() {
	robo := Robo{}
	fmt.Println(robo.Deposit(1, 500))
	fmt.Println(robo.Withdraw(1, 500))
}
