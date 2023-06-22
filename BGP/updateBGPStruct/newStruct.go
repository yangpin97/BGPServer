package updateBGPStruct

import (
	BGPStruct "bgp/DataStruct"
	"bgp/NetCount"
	"bufio"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type ASDetail struct {
	NetMask uint32
	ASPath  []uint32
	Net     uint32
	Type    uint8
	AsSet   []uint32
}

type Result struct {
	*ASDetail
	NetStr    string
	Country   string
	ASInfo    string
	OriginAS  string
	NetPrefix string
}
type MainData struct {
	TotalData     []*ASDetail
	MaskRouteList map[int][]*ASDetail
	ASdata        map[uint32]string
}

func getASName() map[string]string {
	resp, _ := http.Get("https://bgp.potaroo.Net/cidr/autnums.html")
	all, _ := io.ReadAll(resp.Body)
	reader := strings.NewReader(string(all))
	newReader := bufio.NewReader(reader)
	asData := make(map[string]string, 1024)
	for {
		line, _, err := newReader.ReadLine()
		if err != nil {
			fmt.Println("over")
			fmt.Println(err.Error())
			break
		}

		r := regexp.MustCompile(`as=(AS\d+)&view.+?</a>(.+)$`)
		str := string(line)

		submatch := r.FindStringSubmatch(str)

		if len(submatch) >= 3 {

			asData[submatch[1]] = submatch[2]

		}

	}
	fmt.Println("get as")
	return asData

}

func (this *MainData) InitMainData() {

	this.MaskRouteList = make(map[int][]*ASDetail)
	for _, detail := range this.TotalData {

		this.MaskRouteList[int(detail.NetMask)] = append(this.MaskRouteList[int(detail.NetMask)], detail)
	}
	for s, _ := range this.MaskRouteList {
		sort.Slice(this.MaskRouteList[s], func(i, j int) bool {
			return this.MaskRouteList[s][i].Net < this.MaskRouteList[s][j].Net
		})
	}
}

func (this *MainData) GetUSABGP() {
	set := make(map[string]bool)

	routerList := make([]*ASDetail, 0, 1024*10)

	output, err := os.ReadFile("/home/BGP/bgpdump.txt")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("BGP OK")
	var ii = 0
	for _, s := range strings.Split(string(output), "\n") {
		data := strings.Split(s, "|")
		if len(data) >= 5 && !strings.Contains(data[5], "::") {
			ii++
			asnTmp := strings.Split(data[len(data)-2], " ")

			asnpath := NetCount.AsPathtoIntList(asnTmp)
			mask := strings.Split(data[5], "/")[1]
			NetMask, _ := strconv.Atoi(mask)
			start, _ := NetCount.NetToInt(data[5])
			if NetMask == 0 {
				continue
			}
			_, ok := set[data[5]]
			if !ok {
				set[data[5]] = true
			} else {
				continue
			}

			kk := &ASDetail{
				NetMask: uint32(NetMask),
				Net:     uint32(start),
				ASPath:  asnpath,
			}
			routerList = append(routerList, kk)

		}
	}
	this.TotalData = routerList

}
func saveToFile(filename string, routerList map[string]*BGPStruct.ASDetail) error {
	create, err := os.OpenFile("./filename", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	_, _ = create.Write([]byte("fasdf"))
	if err != nil {
		fmt.Println("this _")
		return err
	}
	return nil
}

func whois(ASN string) string {
	cmd := exec.Command("sh", "-c", "whois "+ASN+"| grep -i Country | tail -n 1")
	// 获取命令输出
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	// 输出命令结果
	result := string(output)
	if result == "" {
		return ""
	}
	// 截取结果中的国家简写
	Country := strings.TrimSpace(strings.Split(result, ":")[1])
	return Country
}

func asNameSave(data map[string]string) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.WriteFile("map.json", jsonData, 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Saved file")
}
func asLoadFile() map[string]string {
	jsonData, err := os.ReadFile("map.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var m map[string]string
	err = json.Unmarshal(jsonData, &m)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return m
}

func (this *MainData) Find(ip string) (index int, data1 *Result) {
	ipStr := NetCount.IpToInt(ip)

	for i := 32; i > 5; i-- {
		netMap, ok := this.MaskRouteList[i]
		if ok {
			search := BinarySearchASDetails(netMap, uint32(ipStr))
			if search != -1 {
				return search, this.ChangFormat(this.MaskRouteList[i][search])
			}
		}

	}
	return -1, nil
}
func (this *MainData) FindNetMatch(net string) (index int, data1 *Result) {
	netTmp := strings.Split(net, "/")
	ipStr := NetCount.IpToInt(netTmp[0])
	mask, _ := strconv.Atoi(netTmp[1])
	netMap, ok := this.MaskRouteList[int(mask)]
	if ok {
		search := BinarySearchASDetails(netMap, uint32(ipStr))
		if search != -1 {
			return search, this.ChangFormat(this.MaskRouteList[int(mask)][search])
		}
	} else {
		return this.Find(netTmp[0])
	}

	return -1, nil
}
func BinarySearchASDetails(slice []*ASDetail, targetIP uint32) int {
	low := 0
	high := len(slice) - 1

	for low <= high {
		mid := low + (high-low)/2
		mask := slice[mid].NetMask

		targetIP = targetIP & (0xffffffff << (32 - mask))

		if slice[mid].Net == targetIP {
			return mid // 返回目标IP在切片中的索引
		} else if slice[mid].Net > targetIP {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return -1 // 如果没有找到目标IP，返回-1
}

func (this *MainData) ZipSave(fileName string) {
	file1, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	file := gzip.NewWriter(file1)
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(this.MaskRouteList)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = file1.Close()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()

}

var wg sync.WaitGroup

func (this *MainData) ZipLoad(fileName string) {
	file1, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	go func() {
		wg.Add(1)
		defer wg.Done()
		this.ASdata = NetCount.AsLoadFile()
	}()
	file, _ := gzip.NewReader(file1)

	//file, err := gzip.NewReader(file1)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&this.MaskRouteList)
	if err != nil {
		fmt.Println("--------", err.Error())

		return
	}
	for s, _ := range this.MaskRouteList {
		sort.Slice(this.MaskRouteList[s], func(i, j int) bool {
			return this.MaskRouteList[s][i].Net < this.MaskRouteList[s][j].Net
		})
	}
	wg.Wait()

}
func (this *MainData) Load(fileName string) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	go func() {
		wg.Add(1)
		defer wg.Done()
		this.ASdata = NetCount.AsLoadFile()
	}()

	defer file.Close()

	//file, err := gzip.NewReader(file1)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&this.MaskRouteList)
	if err != nil {
		fmt.Println("--------", err.Error())

		return
	}
	for s, _ := range this.MaskRouteList {
		sort.Slice(this.MaskRouteList[s], func(i, j int) bool {
			return this.MaskRouteList[s][i].Net < this.MaskRouteList[s][j].Net
		})
	}
	wg.Wait()

}
func (this *MainData) ChangFormat(data *ASDetail) *Result {
	ASPathTmp := data.ASPath
	foundAS := ASPathTmp[len(ASPathTmp)-1]
	ASInfo, ok := this.ASdata[foundAS]
	detail := ""
	country := ""
	if ok {
		splitTmp := strings.Split(ASInfo, ",")
		detail = splitTmp[0]
		country = strings.TrimSpace(splitTmp[len(splitTmp)-1])
	}
	result := data
	netStr := NetCount.NumToNetStr(result.Net, uint32(result.NetMask))
	netPrefix := strings.Split(netStr, "/")[0]
	originAS := strconv.Itoa(int(data.ASPath[len(data.ASPath)-1]))
	return &Result{
		ASDetail:  result,
		Country:   country,
		ASInfo:    detail,
		NetStr:    netStr,
		OriginAS:  originAS,
		NetPrefix: netPrefix,
	}
}

func (this *MainData) LoadChinaBGPFile() {
	ChinaBGPList := make([]*ASDetail, 0, 1024*10)
	go func() {
		wg.Add(1)
		defer wg.Done()
		this.ASdata = NetCount.AsLoadFile()
	}()
	file, err := os.ReadFile("newBGP.dat")
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, line := range strings.Split(string(file), "\n") {

		lineData := strings.Split(line, ":")
		net := lineData[0]
		if len(line) > 1 {
			ASPathTmp := strings.Split(lineData[1], " ")
			maskTmp := strings.Split(net, "/")[1]
			mask, _ := strconv.Atoi(maskTmp)
			ASPath := NetCount.AsPathtoIntList(ASPathTmp)

			start, _ := NetCount.NetToInt(net)
			ChinaBGPList = append(ChinaBGPList, &ASDetail{
				NetMask: uint32(mask),
				Net:     uint32(start),
				ASPath:  ASPath,
			})
		}

	}
	this.TotalData = ChinaBGPList
	this.InitMainData()
}

func (this *MainData) Format() {
	ChinaBGPList := make([]*ASDetail, 0, 1024*10)
	go func() {
		wg.Add(1)
		defer wg.Done()
		this.ASdata = NetCount.AsLoadFile()
	}()
	file, err := os.ReadFile("/home/BGP/newBGPfix.txt")
	if err != nil {
		fmt.Println(err.Error())
	}
	str := string(file)
	reader := strings.NewReader(str)
	newReader := bufio.NewReader(reader)
	asSetRe, _ := regexp.Compile(`{(.*?)}`)
	for {
		line, _, err := newReader.ReadLine()
		if err != nil {
			break
		}
		fields := strings.Fields(string(line))
		if len(fields) > 5 {
			routeTypeTmp := fields[0]
			var routeType uint8
			switch routeTypeTmp {
			case "?":
				routeType = 2
			case "e":
				routeType = 1
			case "i":
				routeType = 0
			}
			net := fields[1]
			maskTmp := strings.Split(net, "/")[1]
			mask, _ := strconv.Atoi(maskTmp)
			var asPath []string
			var asSets []string
			start, _ := NetCount.NetToInt(net)
			if strings.Contains(fields[len(fields)-1], "}") {
				asSet := asSetRe.FindStringSubmatch(string(line))
				asSets = strings.Split(asSet[1:][0], " ")
				asPath = fields[5 : len(fields)-len(asSets)]
			} else {
				asPath = fields[5:]
				asSets = []string{}
			}
			ChinaBGPList = append(ChinaBGPList, &ASDetail{
				NetMask: uint32(mask),
				Net:     uint32(start),
				ASPath:  NetCount.AsPathtoIntList(asPath),
				AsSet:   NetCount.AsPathtoIntList(asSets),
				Type:    routeType,
			})
		}

	}
	this.TotalData = ChinaBGPList
	this.InitMainData()

}
