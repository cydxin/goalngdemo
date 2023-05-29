package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
	"sync"
)

type Config struct {
	ImageUploadUrl string    `json:"image_upload_url" yaml:"image_upload_url"` //图片上传地址
	GoodsUploadUrl string    `json:"goods_upload_url" yaml:"goods_upload_url"` //商品信息上传地址
	PageMax        int       `json:"page_max" yaml:"page_max"`                 //最大爬取页数
	ListUrl        []UrlInfo `json:"list_url" yaml:"list_url"`                 //列表url
}

type UrlInfo struct {
	Url          string `json:"url" yaml:"url"`
	CategoryName string `json:"category_name" yaml:"category_name"`
}

var (
	config        Config
	wg            sync.WaitGroup
	loggerSuccess *log.Logger
	loggerError   *log.Logger
	fSucc         *os.File
	fErr          *os.File
	err           error
	ctx           context.Context
)

func init() {
	os.Mkdir("log", 0666)
	fileErrName := fmt.Sprintf("%s/error.log", "log")
	//exists, err := PathExists(fileErrName)
	fErr, err := os.OpenFile(fileErrName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	loggerError = log.New(fErr, "[INFO] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	fileSuccName := fmt.Sprintf("%s/success.log", "log")
	fSucc, err = os.OpenFile(fileSuccName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	// 通过New方法自定义Logger，New的参数对应的是Logger结构体的output, prefix和flag字段
	loggerSuccess = log.New(fSucc, "[INFO] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		loggerSuccess.Println("读取配置文件错误：", err)
		return
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		loggerSuccess.Println("解析配置文件出错")
		return
	}

}

func main() {
	//GetOrderInfo("/dp/B09XXG6MFV/ref=sr_1_22?qid=1667455327&s=kitchen&sr=1-22")
	//return
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // debug使用
		chromedp.UserAgent(ua),           //自定义ua
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	//GetOrderInfo("/dp/B09XXG6MFV/ref=sr_1_22?qid=1667455327&s=kitchen&sr=1-22", "image/B09XXG6MFV", "", "B09XXG6MFV", 1, loggerSuccess, loggerError, ctx)
	//return
	pool := NewPool()
	pool.run()
	for _, urlInfo := range config.ListUrl {
		wg.Add(1)
		pool.JobsChannel <- urlInfo
	}
	wg.Wait()
	loggerSuccess.Println("全部抓取完毕")
	defer fSucc.Close()
	defer fErr.Close()
}

func (p *Pool) Order(urlInfo UrlInfo) {
	page := 1
	defer wg.Done()
	for {
		if page > config.PageMax {
			loggerSuccess.Printf("爬取页数超过限制，停止爬取,url:%s ,page:%d\n", urlInfo.Url, page)
			return
		}
		request, err := GetOrderList(urlInfo.Url, page, ctx)
		if err != nil {
			loggerError.Printf("err:%s", err)
			page++
			continue
		}
		//data := strData[5 : len(strData)-3]
		reader, err := goquery.NewDocumentFromReader(strings.NewReader(request))

		next, _ := reader.Find("#search > div.s-desktop-width-max.s-desktop-content.s-opposite-dir.sg-row > div.s-matching-dir.sg-col-16-of-20.sg-col.sg-col-8-of-12.sg-col-12-of-16 > div > span.rush-component.s-latency-cf-section > div.s-main-slot.s-result-list.s-search-results.sg-row > div:nth-child(27) > div > div > span ").Html()
		if len(next) != 0 {
			//loggerError.Printf("超过了商品最大页数 \npage=%d\nurl=%s\nhtml:%s", page, urlInfo.Url, request)
			loggerError.Printf("pageHtml:%s", page, urlInfo.Url, request)

		}
		reader.Find("#search > div.s-desktop-width-max.s-desktop-content.s-opposite-dir.sg-row > div.s-matching-dir.sg-col-16-of-20.sg-col.sg-col-8-of-12.sg-col-12-of-16 > div > span.rush-component.s-latency-cf-section > div.s-main-slot.s-result-list.s-search-results.sg-row > div").EachWithBreak(func(i int, selection *goquery.Selection) bool {
			asin, ok := selection.Attr("data-asin")
			if ok {
				var infoUrl string
				var result bool
				infoUrl, result = selection.Find(".s-product-image-container > div > span > a").Attr("href")
				//fmt.Println(infoUrl, result)
				if !result {
					infoUrl, result = selection.Find(".s-product-image-container > span > a").Attr("href")
				}
				if result {
					filePath := "image/" + asin
					GetOrderInfo(infoUrl, filePath, urlInfo.CategoryName, asin, 1, loggerSuccess, loggerError, ctx)
				} else {
					htmlstr, _ := selection.Html()
					loggerError.Printf("未获取到商品路径,url:%s\n html:%s", urlInfo.Url, htmlstr)
				}
			}
			return true
		})
		/*for _, value := range strData {
			var htmlData []interface{}
			err = json.Unmarshal([]byte(value), &htmlData)
			if err != nil {
				loggerError.Printf("json Unmarshal err:%s value:%s", err, value)
				continue
			}
			if _, ok := htmlData[2].(map[string]interface{})["asin"]; !ok {
				continue
			}
			if _, ok := htmlData[2].(map[string]interface{})["html"]; !ok {
				continue
			}
			html := htmlData[2].(map[string]interface{})["html"].(string)
			key := htmlData[2].(map[string]interface{})["asin"].(string)
			if len(key) == 0 {
				loggerError.Println("asin 不存在")
				continue
			}
			filePath := "image/" + key
			//fmt.Println(html)
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				loggerError.Printf("documentFromReader err:%s", err)
				continue
			}
			//fmt.Println(html)
			//imageUrl, result := doc.Find(".s-image").Attr("src")
			//infoUrl, result := doc.Find(fmt.Sprintf("#search > div.s-desktop-width-max.s-desktop-content.s-opposite-dir.sg-row > div.s-matching-dir.sg-col-16-of-20.sg-col.sg-col-8-of-12.sg-col-12-of-16 > div > span.rush-component.s-latency-cf-section > div.s-main-slot.s-result-list.s-search-results.sg-row > div:nth-child(%d) > div > div > div > div > div > div > div.s-product-image-container.aok-relative.s-image-overlay-grey.s-text-center.s-padding-left-small.s-padding-right-small.puis-spacing-micro.s-height-equalized > span > a", index+1)).Attr("href")
			var infoUrl string
			var result bool

			infoUrl, result = doc.Find(".s-product-image-container > div > span > a").Attr("href")
			//fmt.Println(infoUrl, result)
			if !result {
				infoUrl, result = doc.Find(".s-product-image-container > span > a").Attr("href")
			}
			if result {
				GetOrderInfo(infoUrl, filePath, urlInfo.CategoryName, key, 1, loggerSuccess, loggerError, ctx)
			} else {
				htmlstr, _ := doc.Html()
				loggerError.Printf("未获取到商品路径,url:%s\n html:%s", urlInfo.Url, htmlstr)
			}
			//fmt.Println(doc.Find(".s-image").Attr("src"))
		}*/
		page++
	}
}

type Pool struct {
	workNum     int
	JobsChannel chan UrlInfo
}

var max = 1

func NewPool() *Pool {
	return &Pool{
		workNum:     max,
		JobsChannel: make(chan UrlInfo),
	}
}

func (p *Pool) run() {
	for i := 1; i <= p.workNum; i++ {
		go p.work()
	}
}
func (p *Pool) work() {
	for task := range p.JobsChannel {
		p.Order(task)
	}
}
