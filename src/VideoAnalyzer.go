package main

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"encoding/xml"
)

/** db connection */
var db *sql.DB

func getDbConnection() *sql.DB {
	db, err := sql.Open("mysql", "testuser:password@/go_lang_test")
	if err != nil {
		panic(err.Error())
	}

	return db
}

func selectNewVideos() *sql.Rows {
	videoNoRows, err := db.Query("SELECT id FROM new_videos ORDER BY post_datetime ASC, id ASC")
	if err != nil && err != sql.ErrNoRows {
		panic(err.Error())
	}

	return videoNoRows
}

func getVideoDetails(videoNo string) {
	url := "http://ext.nicovideo.jp/api/getthumbinfo/sm"
	res, err := http.Get(url + videoNo)
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err.Error())
	}
//	fmt.Println(string(body))

	videoXml := Thumb{Thumb: Video{Title: "", Description: "", ThumbnailUrl: "", Length: "", ViewCounter: "", CommentNum: "", MylistCounter: "", Tags: nil}, Status: ""}
	parseErr := xml.Unmarshal([]byte(body), &videoXml)
	if parseErr != nil {
		panic(parseErr.Error())
	}

	fmt.Println("status: ", videoXml.Status)
	if videoXml.Status != "fail" {
		fmt.Println("title: ", videoXml.Thumb.Title)
		fmt.Println("domain: ", videoXml.Thumb.Tags[0].Domain)
		fmt.Println("tags: ", videoXml.Thumb.Tags[0].Tag)
	}
}

type Thumb struct {
	Thumb Video `xml:"thumb"`
	Status string `xml:"status,attr"`
}

type Video struct {
	Title string `xml:"title"`
	Description string `xml:"description"`
	ThumbnailUrl string `xml:"thumbnail_url"`
	Length string `xml:"length"`
	ViewCounter string `xml:"view_counter"`
	CommentNum string `xml:"comment_num"`
	MylistCounter string `xml:"mylist_counter"`
	Tags []Tags `xml:"tags"`
}

type Tags struct {
	Domain string `xml:"domain,attr"`
	Tag []string `xml:"tag"`
}

func main() {
	db = getDbConnection()
	defer db.Close()

//	getVideoDetails("24512379")

	videoNoRows := selectNewVideos()
	for videoNoRows.Next() {
		var videoNo string
		videoNoRows.Scan(&videoNo)

		fmt.Println("")
		fmt.Println("videoNo : ", videoNo)

		getVideoDetails(videoNo)
	}
}
