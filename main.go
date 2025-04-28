package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type HuelNews struct {
	Id    uint   `json:"id" gorm:"primarykey"`
	Title string `json:"title"`
	Date  string `json:"date"`
}

// 数据库连接初始化
var mydb *gorm.DB

func initDB() error {
	var err error
	dsn := "host=localhost user='c.c' password='' dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	fmt.Printf("dsn: %s\n", dsn)
	mydb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info), //设置全局日志
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	// 自动创建数据库
	if err := mydb.AutoMigrate(&HuelNews{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}

func main() {
	err := initDB()
	if err != nil {
		panic(err)
	}
	ReqUrl := "https://www.huel.edu.cn/index/xydt2.htm"
	ch := make(chan bool)
	for i := 1; i < 72; i++ {
		go func(i int) {
			fmt.Printf("开始爬取第%d页\n", i)
			Spider(ReqUrl, ch)
			fmt.Println("======第", i, "页爬取完成======")
			ReqUrl = "https://www.huel.edu.cn/index/xydt2/" + strconv.Itoa(i) + ".htm"
		}(i)
	}
	for i := 1; i < 72; i++ {
		<-ch
	}
	fmt.Println("全部爬取成功")
}

func Spider(url string, ch chan bool) {
	//1.发送请求
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	//2.解析网页
	DetailDoc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}
	//3.获取节点信息

	//body > div.inner.listbg > div > div.inner_right > div.zn_list > ul > li:nth-child(1)
	//body  > div.inner.listbg > div > div.inner_right > div.zn_list > ul > li:nth-child(1)
	//body > div.inner.listbg > div > div.inner_right > div.zn_list > ul > li:nth-child(3)
	//body > div.inner.listbg > div > div.inner_right > div.zn_list > ul > li:nth-child(1) > a
	//body > div.inner.listbg > div > div.inner_right > div.zn_list > ul > li:nth-child(1) > a > span
	DetailDoc.Find("body > div.inner.listbg > div > div.inner_right > div.zn_list > ul > li").
		Each(func(i int, s *goquery.Selection) {
			title := s.Find("a").Text()
			date := s.Find("a > span").Text()
			data := HuelNews{
				Title: title,
				Date:  date,
			}
			fmt.Printf("正在读取第 %d 个元素", i)
			fmt.Println(data)
			//4.保存内容
			// 保存到数据库
			if err := mydb.FirstOrCreate(&data).Error; err != nil {
				fmt.Printf("保存第 %d 个元素失败: %v\n", i, err)
			} else {
				fmt.Printf("保存第 %d 个元素成功\n", i)
			}
		})

	if ch != nil {
		ch <- true
	}
}

//正则表达式
//func divideInfo()
