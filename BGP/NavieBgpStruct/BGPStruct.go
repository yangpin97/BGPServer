package NavieBgpStruct

type BGP struct {
	Timestamp     float64 `json:"timestamp"`
	Peer          string  `json:"peer"`
	PeerAsn       string  `json:"peer_asn"`
	Id            string  `json:"id"`
	Host          string  `json:"host"`
	Type          string  `json:"type"`
	Path          []int   `json:"path"`
	Community     [][]int `json:"community"`
	Origin        string  `json:"origin"`
	Med           int     `json:"med"`
	Announcements []struct {
		NextHop  string   `json:"next_hop"`
		Prefixes []string `json:"prefixes"`
	} `json:"announcements"`
	Withdrawals []string
}
