package modules

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"sync"
)

// idx = Number days previous the present day.
// 0 means today, 1 means yesterday
// n = Number of images previous the day given by idx
// mkt = Bing Market Area

const (
	mainLink = "https://www.bing.com"
)

type Images struct {
	XMLName xml.Name `xml:"images"`
	Images  []Image  `xml:"image"`
}

type ImageParameters struct {
	Name        string
	Url         string
	Description string
	Filepath    string
}

type Image struct {
	XMLName   xml.Name `xml:"image"`
	Startdate string   `xml:"startdate"`
	Url       string   `xml:"url"`
	Copyright string   `xml:"copyright"`
}

type UrlBing struct {
	url    string
	format string
	idx    int
	n      int
	mkt    string
}

//xmlLink = "https://www.bing.com/HPImageArchive.aspx?format=xml&idx=0&n=7&mkt=en-US"

func MakeUrlString(u *UrlBing) string {

	fields := reflect.TypeOf(*u)
	values := reflect.ValueOf(*u)

	num := fields.NumField()
	var link = u.url
	for i := 1; i < num; i++ {
		field := fields.Field(i)
		value := values.Field(i)
		link += fmt.Sprintf("%s=%v&", field.Name, value)
	}
	return link[:len(link)-1]
}

func parseXML(mutex *sync.Mutex, byteValue chan []byte, u *UrlBing) {
	//defer wg.Done()
	client := &http.Client{}

	mutex.Lock()
	var xmlLink = MakeUrlString(u)
	req, err := http.NewRequest(
		"GET", xmlLink, nil,
	)
	u.idx = 7
	u.n = 8

	mutex.Unlock()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %d, %s", resp.StatusCode, resp.Status)
	}

	var byteVal []byte
	byteVal, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		fmt.Println(e)
		return
	}

	byteValue <- byteVal

}

func GetImageXML(imgParameters *map[string]*ImageParameters) error {
	var byteValue = make(chan []byte)

	//На сайте bing всего доступно 15 картинок
	//Чтобы выцепить все - отправляем 2 запроса
	//idx > 7 не меняет ответ на гет-запрос
	//Максимальное количество картинок в одном запросе n <= 8

	var u = &UrlBing{
		url:    "https://www.bing.com/HPImageArchive.aspx?",
		format: "xml",
		idx:    0,
		n:      7,
		mkt:    "en-US",
	}
	//var wg sync.WaitGroup
	//wg.Add(2)       // в группе две горутины
	const N = 2
	var mutex sync.Mutex
	for i := 0; i < N; i++ {
		go parseXML(&mutex, byteValue, u)
	}

	var images [N]Images

	//wg.Wait()

	for i := range images {
		_ = xml.Unmarshal(<-byteValue, &images[i])
	}

	for i := range images {
		for j := 0; j < len(images[i].Images); j++ {
			var name = images[i].Images[j].Startdate
			if _, exist := (*imgParameters)[name]; !exist {
				var url = images[i].Images[j].Url
				var description = images[i].Images[j].Copyright

				(*imgParameters)[name] = &ImageParameters{}
				(*imgParameters)[name].Name = name

				(*imgParameters)[name].Description = description

				fmt.Println("Image Startdate: " + (*imgParameters)[name].Name)
				fmt.Println("Image Description: " + (*imgParameters)[name].Description)

				re := regexp.MustCompile("([^&]+)")
				match := re.FindStringSubmatch(url)
				downloadLink := mainLink + match[0]
				(*imgParameters)[name].Url = downloadLink
				fmt.Println("Image Url: " + (*imgParameters)[name].Url)
			}
		}
	}
	return nil
}
