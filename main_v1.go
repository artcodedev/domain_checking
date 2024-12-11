package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dex35/smsru"
)

type New_mass struct {
	urls Urls
}

type Conf_s struct {
	Phone        string `json:"phone"`
	Path_to_file string `json:"path_to_file"`
	Tiker        bool   `json:"tiker"`
	Time_sec     int    `json:"time_sec"`
	Token        string `json:"token"`
}

type Urls struct {
	Urls []string `json:"urls"`
}

// if the domain is not up to date or there is another error
func Error_connect_or_bun(u []string, file_name string) {
	data := map[string]interface{}{
		"urls": u,
	}
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile(file_name, file, 0644)
}

// checking one address for relevance
func check_mass_urls(u string) bool {
	res, err := http.Get(u)
	if err != nil {
		return false
	}
	fmt.Println(res.StatusCode)
	if res.StatusCode == 200 {
		return true
	}
	return false
}

// reading a file from the config with all addresses
func read_and_parse_main_file(u string) Urls {
	jsonFile, err := os.Open(u)
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("Файл успешно открыт %s", u)
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var urls_mass Urls
	json.Unmarshal(byteValue, &urls_mass)
	return urls_mass
}

// read conf.json file
func read_conf() []string {
	jsonFile, err := os.Open("conf.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	fmt.Println("Файл успешно открыт conf.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var json_conf_s Conf_s
	json.Unmarshal(byteValue, &json_conf_s)
	var conf_info = []string{}
	conf_info = append(conf_info, string(json_conf_s.Phone))
	conf_info = append(conf_info, string(json_conf_s.Path_to_file))
	conf_info = append(conf_info, fmt.Sprintf("%v", json_conf_s.Tiker))
	conf_info = append(conf_info, strconv.Itoa(json_conf_s.Time_sec))
	conf_info = append(conf_info, string(json_conf_s.Token))
	return conf_info
}

// removing an element from an array by index
func removeByIndex(array []string, index int) []string {
	return append(array[:index], array[index+1:]...)
}

func send_sms(n_url string, ph string, token string) {
	smsClient := smsru.CreateClient(token)
	// send message
	sms := smsru.CreateSMS(ph, "Проверте домен: "+n_url)
	sendedsms, err := smsClient.SmsSend(sms)
	if err != nil {
		log.Panic(err)
	} else {
		log.Printf("Статус запроса: %s, Статус-код выполнения: %d (%s), Баланс: %f", sendedsms.Status, sendedsms.StatusCode, smsru.GetErrorByCode(sendedsms.StatusCode), sendedsms.Balance)
		log.Printf("Сообщение: %s, Статус-код выполнения: %d (%s), Идентификатор: %s, Описание ошибки: %s", sendedsms.Sms[ph].Status, sendedsms.Sms[ph].StatusCode, smsru.GetErrorByCode(sendedsms.StatusCode), sendedsms.Sms[ph].SmsId, sendedsms.Sms[ph].StatusText)
	}
}

// main
func main() {
	fmt.Println("start")

	data_conf := read_conf()

	var urls_mass Urls = read_and_parse_main_file(data_conf[1])
	var n_mass = []string{}
	var stat bool = false
	var error_urls = []string{}

	if len(urls_mass.Urls) != 0 {
		for i := 0; i < len(urls_mass.Urls); i++ {
			u := string(urls_mass.Urls[i])
			log.Printf("Тестовый запрос, домен: %s", u)
			if !check_mass_urls(u) {
				log.Printf("Найден не актуальный адрес или не возможно подключиться!\n<------ %s ------>\n", string(urls_mass.Urls[i]))
				error_urls = append(error_urls, string(urls_mass.Urls[i]))
				stat = true
			} else {
				n_mass = append(n_mass, string(urls_mass.Urls[i]))
			}
		}
		if stat {
			n_str := strings.Join(error_urls, ", ")
			send_sms("["+n_str+"]", data_conf[0], data_conf[4])
			Error_connect_or_bun(n_mass, data_conf[1])
		}
	}

	tiker_s, err := strconv.ParseBool(data_conf[2])
	if err != nil {
		fmt.Println(err)
	}
	if tiker_s {
		t, _ := strconv.Atoi(data_conf[3])
		time.Sleep(time.Duration(t) * time.Second)
		main()
	}

}
