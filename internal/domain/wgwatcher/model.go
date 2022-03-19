package wgwatcher

type Usage struct {
	Peer []Peer
}

type Peer struct {
	Name                string
	LatestHandshakeUnix int64
	TransferRx          int64
	TransferTx          int64
}

type iniConf struct {
	Peer iniPeer `ini:"Peer"`
}

type iniPeer struct {
	PresharedKey string `ini:"PresharedKey"`
}
