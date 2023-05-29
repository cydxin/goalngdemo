package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/PuerkitoBio/goquery"
	pages "github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"html"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

func GetOrderList(urlList string, page int, ctx context.Context) (string, error) {
	//url := strings.Replace(urlList, "?", "/query?", 1) + fmt.Sprintf("&page=%d&ref=sr_pg_%d", page, page)
	url := urlList + fmt.Sprintf("&page=%d&ref=sr_pg_%d", page, page)
	/*head := map[string]string{
		"connection":   "keep-alive",
		"content-type": "application/json",
		"origin":       "https://www.amazon.cn",
		"cookie":       "session-id=259-0325592-8049133; i18n-prefs=INR; ubid-acbin=259-2814338-7242707; session-token=\"+tKcW27EBRIhP3U92C2I6o6CkTTRkKf7IMM7hBF/81PEf5e2gHFL7qcq9SUn63HXCdyYvKqpPTbSXhfw4/cvJP8O4aBEHftKusWAebW8wiEuZaQLLzAM/B2SjHIA0DO5VDhVnjAQokuHkeQPdH8HMEWHOAxIGKZHhfC0x+KtlF4Q2BVFEcEcIhUwJKooIVhF8n0MDK06d+RCIIidOGOTljlmiTEr52ihULmhK3ZWHmI=\"; csm-hit=tb:SKNMW3Z8Z4ABPZVGGSZS+s-SKNMW3Z8Z4ABPZVGGSZS|1668569276742&t:1668569276742&adb:adblk_no; session-id-time=2082758401l",
		//"referer":                      urlList + fmt.Sprintf("&page=%d", page),
		//"user-agent":                   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		"x-amazon-rush-fingerprints":   "AmazonRushAssetLoader:7011D23E8DA35E2A839DB41088018EF427C050D5|AmazonRushFramework:30C682DA85D763052D2443DFCDF3F2DB5B47EB78|AmazonRushRouter:1D7F85C9C471A133FEDA0A6DE6E803E79C1DAC09",
		"x-amazon-s-mismatch-behavior": "ABANDON",
		"x-amazon-s-swrs-version":      "B36823DCCD688DBD7C9B0BF7857BD5FE,D41D8CD98F00B204E9800998ECF8427E",
		"x-requested-with":             "XMLHttpRequest",
	}
	body := map[string]string{
		"customer-action": "pagination",
	}
	bodyStr, _ := json.Marshal(body)
	result, err := Request(url, string(bodyStr), head)
	if err != nil {
		return nil, err
	}
	strData := strings.Split(string(result), "&&&")
	if len(strData) < 5 {
		loggerError.Printf("链接地址有误,page:%d\nurl:%s\nresult:%s", page, url, result)
	}*/
	var request string
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := pages.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(`#nav-logo-sprites`, chromedp.ByQuery),
		chromedp.OuterHTML("body", &request, chromedp.ByQuery),
		//chromedp.JavascriptAttribute("")
	)
	//fmt.Println(request)
	if err != nil {
		return "", err
	}
	if len(request) == 0 {
		return "", errors.New(fmt.Sprintf("链接地址有误,page:%d\nurl:%s\nresult:%s", page, url, request))
	}
	//reader, err := goquery.NewDocumentFromReader(strings.NewReader(request))
	return request, nil
}

func Request(url, body string, head map[string]string) ([]byte, error) {
	client := &http.Client{}
	var req *http.Request
	var err error
	if len(body) == 0 {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	} else {
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))
	}
	if err != nil {
		fmt.Println(err)
	}
	rand.Seed(time.Now().UnixNano())
	req.Header.Set("User-Agent", browser.Random())
	for key, value := range head {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
func RequestFormData(url string, buff *bytes.Buffer, head map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	/*w.WriteField("title", title)
	w.WriteField("image", image)
	w.WriteField("images", orderImgStr)
	w.WriteField("content", context)
	w.WriteField("original_price", originalPrice)
	w.WriteField("source_url", sourceUrl)
	w.WriteField("category_name", categoryName)*/
	client := &http.Client{}
	var req *http.Request
	var err error
	req, err = http.NewRequest(http.MethodPost, url, buff)
	req.Header.Set("Context-Type", w.FormDataContentType())
	if err != nil {
		fmt.Println(err)
	}
	for key, value := range head {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
func GetOrderInfo(urlPath, filePath, categoryName, key string, num int, logs, loggerError *log.Logger, ctx context.Context) {
	url := "https://www.amazon.in" + urlPath
	//url := "https://www.amazon.in/Sukkhi-Lavish-Plated-Necklace-SKR69591/dp/B08S3MSGT9/ref=sr_1_1?qid=1668574000&s=jewelry&sr=1-1"
	//fmt.Println(path, "path")
	//url := "https://www.amazon.com/dp/B07VHZ41L8?ref_=nav_em__k_ods_ha_ta_0_2_4_6"
	/*head := map[string]string{
		"Cookie":     "session-id=259-0325592-8049133; i18n-prefs=INR; ubid-acbin=259-2814338-7242707; session-token=\"+tKcW27EBRIhP3U92C2I6o6CkTTRkKf7IMM7hBF/81PEf5e2gHFL7qcq9SUn63HXCdyYvKqpPTbSXhfw4/cvJP8O4aBEHftKusWAebW8wiEuZaQLLzAM/B2SjHIA0DO5VDhVnjAQokuHkeQPdH8HMEWHOAxIGKZHhfC0x+KtlF4Q2BVFEcEcIhUwJKooIVhF8n0MDK06d+RCIIidOGOTljlmiTEr52ihULmhK3ZWHmI=\"; csm-hit=tb:SKNMW3Z8Z4ABPZVGGSZS+s-SKNMW3Z8Z4ABPZVGGSZS|1668569276742&t:1668569276742&adb:adblk_no; session-id-time=2082758401l",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
		//"Referer":    "https://www.amazon.com/",
		"cache-control":             "max-age=0",
		"device-memory":             "8",
		"upgrade-insecure-requests": "1",
		"downlink":                  "2.95",
		"dpr":                       "2",
		"ect":                       "4g",
		"rtt":                       "250",
		"sec-ch-device-memory":      "8",
	}
	request, err := Request(url, "", head)
	if err != nil {
		loggerError.Printf("Request：%s", err)
		return
	}

	//fmt.Println(string(request))
	reader, err := goquery.NewDocumentFromReader(bytes.NewReader(request))
	if err != nil {
		loggerError.Printf("goquery NewDocumentFromReader:%s", err)
		return
	}*/
	var request string
	chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &request),
		//chromedp.JavascriptAttribute("")
	)

	reader, err := goquery.NewDocumentFromReader(strings.NewReader(request))
	//var text string
	//价格
	originalPrice := reader.Find("#corePriceDisplay_desktop_feature_div > div.a-section.a-spacing-none.aok-align-center > span.a-price.aok-align-center.reinventPricePriceToPayMargin.priceToPay > span:nth-child(2) > span.a-price-whole").Text() + reader.Find("#corePriceDisplay_desktop_feature_div > div.a-section.a-spacing-none.aok-align-center > span.a-price.aok-align-center.reinventPricePriceToPayMargin.priceToPay > span:nth-child(2) > span.a-price-fraction").Text()

	//fmt.Println(reader.Html())
	//无价格停止
	if len(originalPrice) == 0 {
		originalPrice = reader.Find("#corePrice_desktop > div > table > tbody > tr:nth-child(2) > td.a-span12 > span.a-price.a-text-price.a-size-medium.apexPriceToPay > span:nth-child(2)").Text()
		if len(originalPrice) == 0 {
			//fmt.Println(reader.Html())
			//fmt.Println("未获取到商品价格")
			originalPrice = reader.Find("#corePrice_desktop > div > table > tbody > tr:nth-child(1) > td.a-span12 > span.a-price.a-text-price.a-size-medium.apexPriceToPay > span.a-offscreen").Text()
			if len(originalPrice) == 0 {
				id, exists := reader.Find("#variation_model > ul > li.swatchSelect").Attr("id")
				if exists {
					arr := strings.Split(reader.Find(fmt.Sprintf("#%s_price > span", id)).Text(), "₹")
					if len(arr) > 1 {
						originalPrice = arr[1]
					}
				}
				if len(originalPrice) == 0 {
					if num < 3 {
						num++
						GetOrderInfo(urlPath, filePath, categoryName, key, num, logs, loggerError, ctx)
					}
					loggerError.Printf("未获取到商品价格\nurl:%s\nhtml: %s", url, request)
					return
				}

			}

		}
	}
	var imagesArr = make([]string, 0)
	if len(strings.Split(string(request), "'initial':")) > 1 {
		imagesObjStr := strings.Trim(strings.Split(strings.Split(string(request), "'initial':")[1], "'colorToAsin'")[0], " ")
		var imagesObj string

		if len(imagesObjStr) >= 3 {
			imagesObj = imagesObjStr[:len(imagesObjStr)-3]
		}
		var colorImagesArr []ColorImages
		_ = json.Unmarshal([]byte(imagesObj), &colorImagesArr)
		for i := 0; i < len(colorImagesArr); i++ {
			if len(colorImagesArr[i].HiRes) != 0 {
				imagesArr = append(imagesArr, colorImagesArr[i].HiRes)
			}
		}
	}

	/*reader.Find("#altImages > ul").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		//text = sel.Text()
		sel.Find("li").EachWithBreak(func(ii int, se *goquery.Selection) bool {

			imageUrl, exists := se.Find("span > span > span > span > img").Attr("src")
			if !exists {
				imageUrl, exists = se.Find("span > span > span > span > span > img").Attr("src")
				//fmt.Println(imageUrl, 777)
				if !exists {
					//fmt.Println(url, imageUrl, exists)
					return false
				}
			}
			//fmt.Println(imageUrl)
			imagesArr = append(imagesArr, imageUrl)
			return true
		})
		return true
	})*/
	/*if len(imagesArr) == 0 {
		fmt.Println(url)
	}*/
	if len(imagesArr) > 2 {
		imagesArr = imagesArr[:len(imagesArr)-1]
	} else {
		imagesArr = make([]string, 0)
		img := strings.Split(strings.Split(strings.Split(string(request), "'colorImages': { 'initial': [{")[1], `}]`)[0], `"main":`)[1:]
		for i := 0; i < len(img); i++ {
			image := strings.Split(strings.Split(img[i], `},"variant"`)[0], ",")
			imagesArr = append(imagesArr, strings.Split(image[len(image)-2], `":[`)[0][1:])
		}
	}
	var orderImgArr = make([]string, 0)
	//imageArr := strings.Split(imagesArr[0], "/")
	//image := imageArr[len(imageArr)-1]
	for _, imageUrl := range imagesArr {
		imageNameArr := strings.Split(imageUrl, "/")
		if len(imageNameArr) >= 1 {
			imageName := imageNameArr[len(imageNameArr)-1]
			if DownloadFile(imageUrl, filePath, imageName) {
				fileResultData, err := RequestFile(filePath + "/" + imageName)
				if err != nil {
					loggerError.Printf("上传文件错误:%s", err)
					return
				}
				fmt.Println(string(fileResultData))
				var fileResult FileResult
				err = json.Unmarshal(fileResultData, &fileResult)
				if err != nil {
					loggerError.Printf("上传图片错误:%s", err)
					continue
				}
				logs.Println("上传图片成功")
				//上传成功的图片
				orderImgArr = append(orderImgArr, fileResult.Data.Url)
				err = os.RemoveAll(filePath)
				if err != nil {
					loggerError.Printf("删除本地缓存图片失败,err:%s", err)
					continue
				}
				logs.Println("删除本地缓存图片成功")
			}
		}

	}
	if len(orderImgArr) == 0 {
		//fmt.Println("未获取到商品图片")
		loggerError.Printf("未获取到商品图片,url:%s", url)
		return
	}
	var content string
	content, _ = reader.Find("#feature-bullets > ul").Html()
	content = html.EscapeString(content)
	//fmt.Println(context)
	//判断原文是否存在
	/*if len(reader.Find("#ags-mt-popover-featurebullets-update").Text()) == 0 {
		reader.Find("#feature-bullets > ul").EachWithBreak(func(i int, selection *goquery.Selection) bool {
			selection.Find("li").EachWithBreak(func(ii int, sele *goquery.Selection) bool {
				context += sele.Find("span").Text()
				return true
			})
			return true
		})
	} else {
		reader.Find("#a-popover-content-3 > div > ul").EachWithBreak(func(i int, selection *goquery.Selection) bool {
			selection.Find("li").EachWithBreak(func(ii int, sele *goquery.Selection) bool {
				context += sele.Find("span").Text()
				return true
			})
			return true
		})
	}*/
	//去除标题前后空格
	var title string
	if len(reader.Find("#ags-mt-popover-title-update").Text()) == 0 {
		title = strings.Trim(reader.Find("#productTitle").Text(), " ")
	} else {
		title = strings.Trim(reader.Find("#a-popover-content-7 > div").Text(), " ")
	}
	content = base64.StdEncoding.EncodeToString([]byte(strings.Trim(content, " ")))
	//商品分类
	//categoryName := reader.Find("#wayfinding-breadcrumbs_feature_div > ul > li:nth-child(7) > span > a").Text()
	sourceUrl := url
	orderImgStr := strings.Join(orderImgArr, ",")
	var image string
	if len(orderImgArr) > 0 {
		image = orderImgArr[0]
	}
	body := fmt.Sprintf("title=%s&image=%s&images=%s&content=%s&original_price=%s&source_url=%s&category_name=%s&key_sn=%s", base64.StdEncoding.EncodeToString([]byte(title)), base64.StdEncoding.EncodeToString([]byte(image)), base64.StdEncoding.EncodeToString([]byte(orderImgStr)), content, originalPrice, base64.StdEncoding.EncodeToString([]byte(sourceUrl)), categoryName, key)
	orderUrl := config.GoodsUploadUrl
	heads := map[string]string{
		//"Content-Type": w.FormDataContentType(),
		"Content-Type": "application/x-www-form-urlencoded",
	}
	resultData, err := Request(orderUrl, body, heads)
	if err != nil {
		loggerError.Printf("request err:%s", err)
		return
	}
	var result Result
	err = json.Unmarshal(resultData, &result)
	fmt.Println(string(resultData))
	if err != nil {
		loggerError.Printf("上传商品信息失败：%s", err)
		return
	}
	time.Sleep(1 * time.Second)
	logs.Println(result.Msg)
}

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Time string      `json:"time"`
	Data interface{} `json:"data"`
}

type FileResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Time string `json:"time"`
	Data struct {
		Filename    string `json:"filename"`
		Filesize    int    `json:"filesize"`
		Imagewidth  int    `json:"imagewidth"`
		Imageheight int    `json:"imageheight"`
		Imagetype   string `json:"imagetype"`
		Mimetype    string `json:"mimetype"`
		Url         string `json:"url"`
		Uploadtime  int    `json:"uploadtime"`
		Storage     string `json:"storage"`
		Sha1        string `json:"sha1"`
		Createtime  int    `json:"createtime"`
		Updatetime  int    `json:"updatetime"`
		Id          string `json:"id"`
		Fullurl     string `json:"fullurl"`
		ThumbStyle  string `json:"thumb_style"`
	} `json:"data"`
}

func DownloadFile(url, pathName, imageName string) bool {
	resp, err := http.Get(url)
	if err != nil {
		loggerError.Printf("下载图片失败,err:%s,url:%s", err, url)
		return false
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		loggerError.Printf("io ReadAll err:%s", err)
		return false
	}
	if _, err := os.Stat(pathName); os.IsNotExist(err) {
		err := os.MkdirAll(pathName, 0777)
		if err != nil {
			loggerError.Printf("os MkdirAll err:%s", err)
			return false
		}
	}

	fileName := pathName + "/" + imageName
	err = os.WriteFile(fileName, data, 0777)
	if err != nil {
		loggerError.Printf("os WriteFile err:%s", err)
		return false
	}
	return true
}

type ColorImages struct {
	HiRes string `json:"hiRes"`
}

func RequestFile(fileName string) ([]byte, error) {
	url := config.ImageUploadUrl
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(fileName)
	defer file.Close()
	part1, errFile1 := writer.CreateFormFile("file", fileName)
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {
		return nil, errFile1
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Cookie", "token=2dc3654a-ebdb-4c63-9882-2fd098378f0a")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://www.apifox.cn)")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}
