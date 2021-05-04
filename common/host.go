package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sbabiv/xml2map"
)

// Host desc
type Host struct {
	Name    string `json:"name"`
	Connect struct {
		Service     string `json:"service"`
		Address     string `json:"address"`
		Port        int    `json:"port"`
		Credentials struct {
			Key      string `json:"key,omitempty"`
			Login    string `json:"login,omitempty"`
			Password string `json:"password,omitempty"`
		} `json:"credentials"`
	} `json:"connect"`
	Type        string   `json:"type"`
	Parent      string   `json:"parent"`
	Os          string   `json:"os"`
	Ips         []string `json:"ips"`
	Comment     string   `json:"comment"`
	Project     string   `json:"project"`
	Domain      string   `json:"domain"`
	Status      string   `json:"status"`
	Ignore      bool     `json:"ignore"`
	Activeusers []string `json:"activeusers"`
	CPU         struct {
		Type  string `json:"type"`
		Count int    `json:"count"`
	} `json:"cpu"`
	RAM   string `json:"ram"`
	Disks []struct {
		Type     string `json:"type"`
		Capacity string `json:"capacity"`
	} `json:"disks"`
	Ports      []int `json:"ports"`
	Containers []struct {
		Name     string   `json:"name"`
		Services []string `json:"services"`
		Ports    []int    `json:"ports"`
	} `json:"containers"`
	Crontabs []struct {
		User      string `json:"user"`
		Frequency string `json:"frequency"`
		Script    string `json:"script"`
	} `json:"crontabs"`
	Firewall    string `json:"firewall"`
	Lastconnect string `json:"lastconnect"`
}

// HostList is holding hosts list
type HostList struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Hosts  []Host `json:"hosts"`
}

// WholeList = list of all hosts list
var WholeList map[string]HostList

// hostMap is used for hostnames hashing for updating the array
var hostsMap map[string]int

// LoadAll loads all json hosts files in a folder
func LoadAll(folder string) {
	WholeList = make(map[string]HostList)
	matches, _ := filepath.Glob(folder + string(os.PathSeparator) + "*.json")
	for _, match := range matches {
		list := LoadSingleFile(match)
		WholeList[list.Name] = list
	}
}

// LoadSingleFile from json file
func LoadSingleFile(filename string) HostList {
	var list HostList
	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &list)
	return list
}

// SaveAll to json filme
func SaveAll(list HostList, folder string) {
	for key, value := range WholeList {
		SaveList(value, folder+string(os.PathSeparator)+strings.ReplaceAll(strings.ToLower(key), " ", "_")+".json")
	}
}

// SaveList to json filme
func SaveList(list HostList, filename string) {
	file, _ := json.MarshalIndent(list, "", " ")
	_ = ioutil.WriteFile(filename, file, 0644)
}

// Init list
func Init() {
	WholeList = make(map[string]HostList)
}

// UpdateFromEsx from esx glpi scan, load all files
func UpdateFromEsx(list HostList) {
	//cmd := exec.Command("sleep", "1")
	fmt.Println("Running GLPI Inventory...")
	//err := cmd.Run()
	//fmt.Printf("Command finished with error: %v", err)
	//if err == nil {
	files, _ := filepath.Glob("*.ocs")
	for _, file := range files {
		updateFromEsxFile(list, file)
	}
	//}

}

// UpdateFromEsxFile from esx fusion inventory scan
func updateFromEsxFile(list HostList, filename string) {
	hostsMap = make(map[string]int)
	for index, element := range list.Hosts {
		hostsMap[element.Name] = index
	}
	xmlFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()
	data, _ := ioutil.ReadAll(xmlFile)
	decoder := xml2map.NewDecoder(strings.NewReader(string(data)))
	result, err := decoder.Decode()

	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		data := result["REQUEST"].(map[string]interface{})["CONTENT"].(map[string]interface{})["VIRTUALMACHINES"]
		//fmt.Printf("%v\n", data)
		datalist := data.([]map[string]interface{})
		fmt.Printf("%v records in ocs file\n", len(datalist))

		for _, element := range datalist {
			val, ok := hostsMap[element["NAME"].(string)]
			if ok {
				fmt.Println("updating", element["NAME"].(string))
				list.Hosts[val].Comment = element["COMMENT"].(string)
				list.Hosts[val].Type = element["VMTYPE"].(string)
				list.Hosts[val].CPU.Type = "VCPU"
				cpu, err := strconv.Atoi(element["VCPU"].(string))
				if err == nil {
					list.Hosts[val].CPU.Count = cpu
				} else {
					list.Hosts[val].CPU.Count = 0
				}
				list.Hosts[val].RAM = element["MEMORY"].(string)
				list.Hosts[val].Status = element["STATUS"].(string)
			} else {
				fmt.Println("adding", element["NAME"].(string))
				//vm := element.(map[string]string)
				var host Host
				host.Name = element["NAME"].(string)
				host.Comment = element["COMMENT"].(string)
				host.Type = element["VMTYPE"].(string)
				host.CPU.Type = "VCPU"
				cpu, err := strconv.Atoi(element["VCPU"].(string))
				if err == nil {
					host.CPU.Count = 0
				} else {
					host.CPU.Count = cpu
				}
				host.RAM = element["MEMORY"].(string)
				host.Status = element["STATUS"].(string)
				list.Hosts = append(list.Hosts, host)
			}
			//fmt.Printf("%v\n", element)
		}
	}
}
