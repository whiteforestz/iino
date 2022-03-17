package wgwatcher

type Usage struct {
	Peer []PeerUsage
}

type PeerUsage struct {
	Name                string
	LatestHandshakeUnix int64
}

type iniConf struct {
	Peer iniPeer `ini:"Peer"`
}

type iniPeer struct {
	PresharedKey string `ini:"PresharedKey"`
}
