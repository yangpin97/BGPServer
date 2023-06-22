package main

import (
	"bgp/NetCount"
	"bgp/updateBGPStruct"
	ASDetail "bgp/updateBGPStruct"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/log"
	"github.com/osrg/gobgp/v3/pkg/server"
	"github.com/sirupsen/logrus"
	apb "google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/ini.v1"
	sdlog "log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var ChinaBGP updateBGPStruct.MainData
var USA_BGP updateBGPStruct.MainData

type Params struct {
	MoreSpecific  bool            `json:"moreSpecific"`
	Host          string          `json:"host"`
	SocketOptions map[string]bool `json:"socketOptions"`
}

type Message struct {
	Type string `json:"type"`
	Data Params `json:"data"`
}

type BGP struct {
	Type string `json:"type"`
	Data struct {
		Timestamp     float64       `json:"timestamp"`
		Peer          string        `json:"peer"`
		PeerAsn       string        `json:"peer_asn"`
		Id            string        `json:"id"`
		Host          string        `json:"host"`
		Type          string        `json:"type"`
		Path          []interface{} `json:"path"`
		Community     [][]uint32    `json:"community"`
		Origin        string        `json:"origin"`
		Aggregator    string        `json:"aggregator"`
		Announcements []struct {
			NextHop  string   `json:"next_hop"`
			Prefixes []string `json:"prefixes"`
		} `json:"announcements"`
		Withdrawals []string `json:"withdrawals"`
		Raw         string   `json:"raw"`
	} `json:"data"`
}
type myLogger struct {
	logger *logrus.Logger
}

func (l *myLogger) Panic(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Panic(msg)
}

func (l *myLogger) Fatal(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Fatal(msg)
}

func (l *myLogger) Error(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Error(msg)
}

func (l *myLogger) Warn(msg string, fields log.Fields) {
	l.logger.WithFields(logrus.Fields(fields)).Warn(msg)
}

func (l *myLogger) Info(msg string, fields log.Fields) {
	//l.logger.WithFields(logrus.Fields(fields)).Info(msg)
}

func (l *myLogger) Debug(msg string, fields log.Fields) {
	//l.logger.WithFields(logrus.Fields(fields)).Debug(msg)
}

func (l *myLogger) SetLevel(level log.LogLevel) {
	//l.logger.SetLevel(logrus.Level(level))
}

func (l *myLogger) GetLevel() log.LogLevel {
	return log.LogLevel(l.logger.GetLevel())
}

var ASdata = NetCount.AsLoadFile()

func DecodePATH(path []interface{}) []uint32 {
	newPath := make([]uint32, 0, 32)
	for _, element := range path {
		switch v := element.(type) {
		case float64:
			newPath = append(newPath, uint32(v))
		case []interface{}:

			//for _, element2 := range v {
			//	switch _ = element2.(type) {
			//	case float64:
			//
			//	}
			//}
		}
	}
	return newPath
}

func ChangFormat(data *BGP) (add, del []*updateBGPStruct.Result) {
	if strings.Contains(data.Data.Peer, ":") || (len(data.Data.Announcements) < 1 && len(data.Data.Withdrawals) < 1) {
		return nil, nil
	}
	addRouters := make([]*updateBGPStruct.Result, 0, 1)
	delRouters := make([]*updateBGPStruct.Result, 0, 1)
	path := DecodePATH(data.Data.Path)
	if len(path) > 0 {

		ASInfo, ok := ASdata[path[len(path)-1]]
		detail := ""
		country := ""
		if ok {
			splitTmp := strings.Split(ASInfo, ",")
			detail = splitTmp[0]
			country = strings.TrimSpace(splitTmp[len(splitTmp)-1])
		}

		for _, prefix := range data.Data.Announcements[0].Prefixes {
			if strings.Contains(prefix, ":") {
				continue
			}
			prefixTmp := strings.Split(prefix, "/")
			mask, _ := strconv.Atoi(prefixTmp[1])
			addRouters = append(addRouters, &updateBGPStruct.Result{
				ASDetail: &ASDetail.ASDetail{
					NetMask: uint32(mask),
					ASPath:  path,
				},
				NetPrefix: prefixTmp[0],
				ASInfo:    detail,
				Country:   country,
			})
		}
	}
	if len(data.Data.Withdrawals) > 0 {
		for _, delPrefix := range data.Data.Withdrawals {
			if strings.Contains(delPrefix, ":") {
				continue
			}
			delprefixTmp := strings.Split(delPrefix, "/")
			mask, _ := strconv.Atoi(delprefixTmp[1])
			delRouters = append(delRouters, &updateBGPStruct.Result{
				ASDetail: &ASDetail.ASDetail{
					NetMask: uint32(mask),
					ASPath:  path,
				},
				NetPrefix: delprefixTmp[0],
			})

		}
	}

	return addRouters, delRouters
}
func Add(s *server.BgpServer, result *updateBGPStruct.Result) {
	nlri, _ := apb.New(&api.IPAddressPrefix{
		Prefix:    result.NetPrefix,
		PrefixLen: result.NetMask,
	})
	a5, _ := apb.New(&api.MultiExitDiscAttribute{
		Med: uint32(100),
	})
	a1, _ := apb.New(&api.OriginAttribute{
		Origin: uint32(result.Type),
	})
	a2, _ := apb.New(&api.NextHopAttribute{
		NextHop: bgpConfig.NextHop,
	})
	a3, _ := apb.New(&api.AsPathAttribute{
		Segments: []*api.AsSegment{
			{
				Type:    2,
				Numbers: result.ASPath,
			},
			{
				Type:    1,
				Numbers: result.AsSet,
			},
		},
	})
	//var a7 *apb.Any
	//if len(result.AsSet) > 0 {
	//	a7, _ = apb.New(&api.AggregatorAttribute{
	//		Asn:     uint32(32),
	//		Address: "x.x.x.x,
	//	})
	//} else {
	//	a7 = nil
	//}

	attrs := []*apb.Any{a1, a2, a3, a5}

	_, err := s.AddPath(context.Background(), &api.AddPathRequest{
		Path: &api.Path{
			Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

}
func Del(s *server.BgpServer, result *updateBGPStruct.Result) {
	a2, _ := apb.New(&api.NextHopAttribute{
		NextHop: bgpConfig.NextHop,
	})

	attrs := []*apb.Any{a2}

	nlri, _ := apb.New(&api.IPAddressPrefix{
		Prefix:    result.NetPrefix,
		PrefixLen: result.NetMask,
	})
	err := s.DeletePath(context.Background(), &api.DeletePathRequest{
		Path: &api.Path{
			Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})
	if err != nil {
		fmt.Println("err")
		fmt.Println("err")
		fmt.Println("err")
		fmt.Println("err")
		fmt.Println("err")
		fmt.Println(err.Error())
		return
	}

}
func Get(s *server.BgpServer) {
	fmt.Println("Start connect to BGP server")
	var dialer *websocket.Dialer
	ws, _, err := dialer.Dial("wss://ris-live.ripe.net/v1/ws/?client=go-example-1", nil)
	if err != nil {
		fmt.Println("dial:", err)
	}
	defer ws.Close()

	params := Params{
		MoreSpecific: true,
		Host:         "rrc21",
		SocketOptions: map[string]bool{
			"includeRaw": true,
		},
	}
	msg := Message{
		Type: "ris_subscribe",
		Data: params,
	}

	message, _ := json.Marshal(msg)
	ws.WriteMessage(websocket.TextMessage, message)
	for {
		_, messageTpye, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			Get(s)
			return
		}
		var bgpData BGP
		err = json.Unmarshal(messageTpye, &bgpData)

		if err != nil {
			fmt.Println(string(messageTpye))
			fmt.Println("json err1", err.Error())
			return
		}
		add, del := ChangFormat(&bgpData)
		for _, route := range add {

			flag := findInCN(route.NetPrefix)

			if flag {

				Add(s, route)
				sdlog.Println("add  ", route.ASPath, route.Country, route.NetPrefix, route.ASInfo)

			}

		}
		for _, route := range del {
			flag := findInCN(route.NetPrefix)

			if flag {
				sdlog.Println("del  ", route.ASPath, route.Country, route.NetPrefix, route.ASInfo)
				Del(s, route)

			}
		}

	}
}

var wg sync.WaitGroup
var COUNTRY string

func findInCN(IP string) bool {
	_, data1 := ChinaBGP.Find(IP)

	if COUNTRY == "ALL" {
		return true
	}
	if COUNTRY == "NCN" {
		if data1.Country != "CN" {
			return true
		} else {
			return false
		}
	}
	if data1 != nil {
		if data1.Country == COUNTRY {
			return true
		} else {
			return false
		}
	}
	//else {
	//	_, result := USA_BGP.Find(IP)
	//	if result != nil {
	//		if result.Country == "CN" {
	//			return true
	//		}
	//	}
	//
	//}
	return false
}

type BGPConfig struct {
	ID           string
	ServerASN    int
	NextHop      string
	ClientIP     string
	ClientASN    int
	UpdateSource string
}

var bgpConfig BGPConfig

func loadIni(configPath string) *BGPConfig {
	cfg, _ := ini.Load(configPath)
	fmt.Println("start load")

	bgpConfig.ID = cfg.Section("server").Key("RouterId").String()
	//
	sAsn, _ := strconv.Atoi(cfg.Section("server").Key("ASN").String())
	bgpConfig.ServerASN = sAsn
	bgpConfig.NextHop = cfg.Section("server").Key("NextHop").String()
	bgpConfig.ClientIP = cfg.Section("peer").Key("IP").String()
	bgpConfig.UpdateSource = cfg.Section("server").Key("UpdateSource").String()
	cAsn, _ := strconv.Atoi(cfg.Section("peer").Key("ASN").String())
	bgpConfig.ClientASN = cAsn

	return &bgpConfig

}

func main() {
	log := logrus.New()
	configFile := flag.String("c", "config.json", "config file")
	userCountry := flag.String("l", "CN", "country")

	flag.Parse()
	*userCountry = strings.ToUpper(*userCountry)
	COUNTRY = *userCountry
	fmt.Println(configFile)
	bgpConfig := loadIni(*configFile)

	s := server.NewBgpServer(server.LoggerOption(&myLogger{logger: log}))
	wg.Add(1)
	go s.Serve()

	// global configuration
	if err := s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			Asn:        uint32(bgpConfig.ServerASN),
			RouterId:   bgpConfig.ID,
			ListenPort: 179, // -1 gobgp won't listen on tcp:
			//UseMultiplePaths: true,
			ListenAddresses: []string{"0.0.0.0"},
			GracefulRestart: &api.GracefulRestart{
				Enabled:             true, // 启用Graceful Restart
				RestartTime:         120,  // 重启时间（以秒为单位）
				NotificationEnabled: true, // 启用Graceful Restart通知
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	// monitor the change of the peer state
	if err := s.WatchEvent(context.Background(), &api.WatchEventRequest{Peer: &api.WatchEventRequest_Peer{}}, func(r *api.WatchEventResponse) {
		if p := r.GetPeer(); p != nil && p.Type == api.WatchEventResponse_PeerEvent_STATE {
			log.Info(p)
		}
	}); err != nil {
		log.Fatal(err)
	}
	// neighbor configuration
	n := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: bgpConfig.ClientIP,
			PeerAsn:         uint32(bgpConfig.ClientASN),
		},
		Transport: &api.Transport{
			LocalAddress: bgpConfig.UpdateSource,
		},
		EbgpMultihop: &api.EbgpMultihop{
			Enabled:     true,
			MultihopTtl: 32, // 设置跳数
		},
	}
	if err := s.AddPeer(context.Background(), &api.AddPeerRequest{

		Peer: n,
	}); err != nil {
		log.Fatal(err)
	}
	// add routes
	// do something useful here instead of exiting
	//USA_BGP.ZipLoad("USA_BGPZip.gob")
	executable, _ := os.Executable()
	dir := filepath.Dir(executable)
	fileDir := filepath.Join(dir, "ChinaBGPZip.gob")
	ChinaBGP.ZipLoad(fileDir)
	for _, k := range ChinaBGP.MaskRouteList {

		for _, data := range k {
			format := ChinaBGP.ChangFormat(data)
			if *userCountry == "CN" {
				if format.Country == "CN" {
					Add(s, format)
				}
			} else if *userCountry == "NCN" {
				if format.Country != "CN" {
					Add(s, format)
				}
			} else if *userCountry == "ALL" {
				Add(s, format)
			} else {
				if format.Country == COUNTRY {
					Add(s, format)
				}

			}

		}
	}
	Get(s)

	wg.Wait()
}

// implement github.com/osrg/gobgp/v3/pkg/log/Logger interface
