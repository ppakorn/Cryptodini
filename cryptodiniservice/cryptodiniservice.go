package cryptodiniservice

type UserID = int

type Coin struct {
	Symbol string
	Amount float64
}

type Portfolio struct {
	Id    int
	Coins []Coin
}

// Dummy data
var port = Portfolio{
	Id: 1,
	Coins: []Coin{
		Coin{Symbol: "BTC", Amount: 0.04},
		Coin{Symbol: "ETH", Amount: 1.5},
		Coin{Symbol: "XRP", Amount: 800},
	},
}

type CryptodiniService interface {
	Adjust(uid UserID, desiredPort *Portfolio)
	GetPort(uid UserID) *Portfolio
}

func Adjust(uid UserID, desiredPort *Portfolio) {
	// TODO: actually call to port service
	port = *desiredPort
}

func GetPort(uid UserID) *Portfolio {
	// TODO: actually call to port service
	return &port
}
