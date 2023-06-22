package BGPStruct

type ASDetail struct {
	Operation string
	Peer      string
	OriginAS  string
	ASDetail  string
	NetMask   uint8
	ASPath    []string
	Country   string
	IpStart   int
	IpEnd     int
	Net       string
}
