package core

type NetID uint8

const (
	MainNet       NetID = 0
	RegressionNet NetID = 1
	TestNet3      NetID = 2
	SimNet        NetID = 3
)

var NetIDToName = []string{
	MainNet:       "mainnet",
	RegressionNet: "regressionnet",
	TestNet3:      "testnet3",
	SimNet:        "simnet",
}

func (t NetID) String() string {
	if int(t) > len(NetIDToName) || int(t) < 0 {
		return "Invalid"
	}
	return NetIDToName[t]
}
