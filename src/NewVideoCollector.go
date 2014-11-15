package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"container/list"
	_ "github.com/go-sql-driver/mysql"
	"github.com/PuerkitoBio/goquery"
	"time"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

/** configuration */
var config Config

type Config struct {
	Db DbConfig `json:"db"`
}

type DbConfig struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Name string `json:"name"`
}

func loadConfig() {
	file, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(file, &config)
}

/** db connection */
var db *sql.DB

func getDbConnection() *sql.DB {
	db, err := sql.Open("mysql", config.Db.User + ":" + config.Db.Pass + "@/" + config.Db.Name)
	if err != nil {
		panic(err.Error())
	}

	return db
}

func selectLastCollectedVideo() (string, string) {
	var videoId string
	var postDateTime string
	err := db.QueryRow("SELECT id, post_datetime FROM new_videos ORDER BY post_datetime DESC, id DESC LIMIT 1").Scan(&videoId, &postDateTime)
	if err != nil && err != sql.ErrNoRows{
		panic(err.Error())
	}

	return videoId, postDateTime
}

func collectNewVideo(endVideoId string, endDateTime string) *list.List {
	var count int
	var limit int

	if endVideoId == "" || endDateTime == "" {
		count = 0
		limit = 101
	}

	videos := list.New()
	next := true;
	for pageNo := 1; next; pageNo++ {
		doc := getSearchResultDoc(pageNo)

		doc.Find(".contentBody.uad:not(.searchUad).video .item").Each(func(_ int, s *goquery.Selection) {
			rawVideoId, _ := s.Attr("data-id")
			videoId := regexp.MustCompile("[0-9]+").FindString(rawVideoId)
			postDatetime := regexp.MustCompile("[ /:]").ReplaceAllString(s.Find(".itemTime .time:not(.new)").Text(), "")
			title, _ := s.Find(".itemTitle a").Attr("title")

			if len(postDatetime) == 10 {
				postDatetime = "20" + postDatetime
			} else if len(postDatetime) == 8 {
				postMonth, _ := strconv.Atoi(postDatetime[0:2])
				nowMonth, _ := strconv.Atoi(fmt.Sprint(time.Now().Month()))
				if nowMonth < postMonth {
					postDatetime = fmt.Sprint(time.Now().AddDate(-1, 0, 0).Year()) + postDatetime
				} else {
					postDatetime = fmt.Sprint(time.Now().Year())+postDatetime
				}
			} else {
				panic("投稿日時の長さがおかしいですよ")
			}

			if videoId == endVideoId || postDatetime < endDateTime || (limit != 0 && limit <= count) {
				next = false;
				return;
			}
			videoMap := map[string]string{"id": videoId, "datetime": postDatetime, "title": title}
			videos.PushBack(videoMap)
			count++
		})
	}

	return videos
}

func getSearchResultDoc(pageNo int) *goquery.Document {
	url := "http://www.nicovideo.jp/tag/%E5%AE%9F%E6%B3%81%E3%83%97%E3%83%AC%E3%82%A4%E5%8B%95%E7%94%BB"
	hash := "?sort=f&page=" + fmt.Sprint(pageNo)

	doc, err := goquery.NewDocument(url + hash)
	if err != nil {
		panic(err.Error())
	}

	return doc
}

func registerNewVideos(videos *list.List) {
	recentlyVideoRows := selectRecentlyVideos()

	for recentlyVideoRows.Next() {
		var recentlyMovieNo string
		recentlyVideoRows.Scan(&recentlyMovieNo)
	}

	videoLoop: for video := videos.Back(); video != nil; video = video.Prev() {

		// すでに登録されている動画はスキップする
		videoObj := video.Value.(map[string]string)
		for recentlyVideoRows.Next() {
			var recentlyMovieNo string
			recentlyVideoRows.Scan(&recentlyMovieNo)
			if videoObj["id"] == recentlyMovieNo {
				continue videoLoop
			}
		}

		stmtIns, stmtErr := db.Prepare("INSERT INTO new_videos (id, title, post_datetime, status) VALUES( ?, ?, ?, ?)")
		if stmtErr != nil {
			panic(stmtErr.Error())
		}
		defer stmtIns.Close()

		fmt.Println(videoObj["id"], " ", videoObj["datetime"], " ", videoObj["title"])
		_, insErr := stmtIns.Exec(videoObj["id"], videoObj["title"], videoObj["datetime"], 0)
		if insErr != nil {
			panic(insErr.Error())
		}
	}
}

func selectRecentlyVideos() *sql.Rows {

	videoIdRows, err := db.Query("SELECT id FROM new_videos WHERE post_datetime = (SELECT MAX(post_datetime) FROM new_videos) ORDER BY post_datetime")
	if err != nil && err != sql.ErrNoRows{
		panic(err.Error())
	}

	return videoIdRows
}

func main() {
	loadConfig()
	db = getDbConnection()
	defer db.Close()

	videos := collectNewVideo(selectLastCollectedVideo())
	registerNewVideos(videos)
}
