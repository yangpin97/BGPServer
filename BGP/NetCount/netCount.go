package NetCount

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func NetToInt(Net string) (start, end int) {
	IP := strings.Split(Net, "/")[0]
	IpStartNum := IpToInt(IP)
	mask := strings.Split(Net, "/")[1]
	maskTmp, _ := strconv.Atoi(mask)
	IpEndNum := IpStartNum + 1<<(32-maskTmp) - 1
	return IpStartNum, IpEndNum
}
func IpToInt(IP string) int {
	IpInlist := strings.Split(IP, ".")
	a, _ := strconv.Atoi(IpInlist[0])
	b, _ := strconv.Atoi(IpInlist[1])
	c, _ := strconv.Atoi(IpInlist[2])
	d, _ := strconv.Atoi(IpInlist[3])
	IpNum := a<<24 | b<<16 | c<<8 | d
	return IpNum
}
func NumToNetStr(ip uint32, mask uint32) string {
	maskStr := strconv.Itoa(int(mask))
	ipSlice := make([]uint32, 4)
	ipSlice[3] = ip >> 24
	ipSlice[2] = ip & 0x00ff0000 >> 16
	ipSlice[1] = ip & 0x0000ff00 >> 8
	ipSlice[0] = ip & 0x000000ff
	var netStr string
	netStr = strconv.Itoa(int(ipSlice[3])) + "." + strconv.Itoa(int(ipSlice[2])) + "." + strconv.Itoa(int(ipSlice[1])) + "." + strconv.Itoa(int(ipSlice[0]))
	netStr = netStr + "/" + maskStr
	return netStr
}

func Whois(ASN string) string {
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

func AsNameSave(data map[uint32]string) {
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

var m map[uint32]string

func AsLoadFile() map[uint32]string {
	if len(m) > 0 {
		return m
	} else {
		fmt.Println("load asName ----------------------------------------------")
		executable, _ := os.Executable()
		dir := filepath.Dir(executable)
		join := filepath.Join(dir, "map.json.gz")
		file1, _ := os.Open(join)
		file, _ := gzip.NewReader(file1)
		jsonData, err := io.ReadAll(file)
		if err != nil {

			return nil
		}

		err = json.Unmarshal(jsonData, &m)
		if err != nil {
			fmt.Println("errrrrrr", err.Error())
			return nil
		}
		return m
	}

}

func AsPathtoIntList(data []string) []uint32 {
	asnList := make([]uint32, 0, 12)
	for _, as := range data {
		if !strings.Contains(as, "{") {
			asn, err := strconv.Atoi(as)
			if err != nil {
				fmt.Println(data)
				fmt.Println(len(data))
				fmt.Println(err.Error())
				return []uint32{}
			}
			asnList = append(asnList, uint32(asn))
		}

	}
	return asnList
}

func GetASNameToMapNum() map[uint32]string {
	resp, _ := http.Get("https://bgp.potaroo.net/cidr/autnums.html")
	all, _ := io.ReadAll(resp.Body)
	reader := strings.NewReader(string(all))
	newReader := bufio.NewReader(reader)
	asData := make(map[uint32]string, 1024)
	for {
		line, _, err := newReader.ReadLine()

		if err != nil {
			fmt.Println("over")
			fmt.Println(err.Error())
			break
		}

		r := regexp.MustCompile(`as=AS(\d+)&view.+?</a>(.+)$`)
		str := string(line)
		submatch := r.FindStringSubmatch(str)

		if len(submatch) >= 3 {
			asn, _ := strconv.Atoi(submatch[1])
			asData[uint32(asn)] = submatch[2]

		}

	}
	fmt.Println("get as")
	return asData

}
