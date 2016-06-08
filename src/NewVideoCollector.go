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
	"strings"
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
	err := db.QueryRow("SELECT id, post_datetime FROM new_videos ORDER BY serial_no DESC LIMIT 1").Scan(&videoId, &postDateTime)
	if err != nil && err != sql.ErrNoRows{
		panic(err.Error())
	}

	return videoId, postDateTime
}

func collectNewVideo(endVideoId string, endDateTime string, videos *list.List) {
	count := 0
	breakCount := 0
	limit := 300
	next := true;

	for pageNo := 1; next; pageNo++ {
		doc := getSearchResultDoc(pageNo)

		doc.Find(".thumb_col_1").Each(func(_ int, s *goquery.Selection) {
			videoLink := s.Find(".watch")
			rawVideoId, _ := videoLink.Attr("href")
			videoId := regexp.MustCompile("[0-9]+").FindString(rawVideoId)
			postDatetime := regexp.MustCompile("[年月日 /:]").ReplaceAllString(s.Find(".thumb_num strong").Text(), "")
			title, _ := videoLink.Attr("title")

			if len(postDatetime) == 12 {
				// NOP
			} else if len(postDatetime) == 10 {
				postDatetime = "20" + postDatetime
			} else if len(postDatetime) == 8 {
				postMonth, _ := strconv.Atoi(postDatetime[0:2])
				nowMonth, _ := strconv.Atoi(fmt.Sprint(time.Now().Month()))
				if nowMonth < postMonth {
					postDatetime = fmt.Sprint(time.Now().AddDate(-1, 0, 0).Year()) + postDatetime
				} else {
					postDatetime = fmt.Sprint(time.Now().Year()) + postDatetime
				}
			} else {
				panic("投稿日時の長さがおかしいですよ")
			}

			if postDatetime < endDateTime {
				breakCount++
			}

			// 読み込み中断判定
			if (endVideoId != "" && 100 <= breakCount) || (limit != 0 && limit <= count) {
				next = false;
				count++
				return;
			}

			isNewVideo := true
			for vi := videos.Front(); vi != nil; vi = vi.Next() {
				viMap := vi.Value.(map[string]string)
				if videoId == viMap["id"] {
					isNewVideo = false
					continue
				}
			}

			if isNewVideo {
				videoMap := map[string]string{"id": videoId, "datetime": postDatetime, "title": title}
				videos.PushBack(videoMap)
			}
			count++
		})
	}
}

func collectNewVideoByCategory(endVideoId string, endDateTime string, tags []string, videos *list.List) {
	count := 0
	breakCount := 0
	limit := 100
	next := true;

	for pageNo := 1; next && pageNo <= 100; pageNo++ {
		// スリープで短時間での連続アクセスを避ける
		time.Sleep(1000 * time.Millisecond)

		doc := getSearchResultDocByCategory(pageNo, tags)

		doc.Find(".contentBody.uad:not(.searchUad).video .item").Each(func(_ int, s *goquery.Selection) {
			rawVideoId, _ := s.Attr("data-id")
			videoId := regexp.MustCompile("[0-9]+").FindString(rawVideoId)
			postDatetime := regexp.MustCompile("[ /:]").ReplaceAllString(s.Find(".itemTime .time:not(.new)").Text(), "")
			title, _ := s.Find(".itemTitle a").Attr("title")

			if len(postDatetime) == 10 {
				postDatetime = "20" + postDatetime
			} else if len(postDatetime) == 8 {
				postMonth, _ := strconv.Atoi(postDatetime[0:2])
				nowMonth := int(time.Now().Month())
				if nowMonth < postMonth {
					postDatetime = fmt.Sprint(time.Now().AddDate(-1, 0, 0).Year()) + postDatetime
				} else {
					postDatetime = fmt.Sprint(time.Now().Year())+postDatetime
				}
			} else {
				panic("投稿日時の長さがおかしいですよ")
			}

			if postDatetime < endDateTime {
				breakCount++
			}

			// 読み込み中断判定
			if (endVideoId != "" && 100 <= breakCount) || (limit != 0 && limit <= count) {
				next = false;
				return;
			}

			for vi := videos.Front(); vi != nil; vi = vi.Next() {
				viMap := vi.Value.(map[string]string)
				if videoId == viMap["id"] {
					continue
				}
			}

			videoMap := map[string]string{"id": videoId, "datetime": postDatetime, "title": title}
			videos.PushBack(videoMap)
			count++
		})
	}
}

func getSearchResultDoc(pageNo int) *goquery.Document {
	url := "http://www.nicovideo.jp/newarrival"
	hash := "?sort=f&page=" + fmt.Sprint(pageNo)

	// スリープで短時間での連続アクセスを避ける
	time.Sleep(300 * time.Millisecond)

	doc, err:= goquery.NewDocument(url + hash)
	if err != nil {
		panic(err.Error())
	}

	return doc
}

func getSearchResultDocByCategory(pageNo int, tags []string) *goquery.Document {
	url := "http://www.nicovideo.jp/tag/" + strings.Join(tags, " or ")
	hash := "?sort=f&order=d&ref=cate_newall&page=" + fmt.Sprint(pageNo)

	// スリープで短時間での連続アクセスを避ける
	time.Sleep(300 * time.Millisecond)

	doc, err:= goquery.NewDocument(url + hash)
	if err != nil {
		panic(err.Error())
	}

	return doc
}

func registerNewVideos(videos *list.List) {
	insertCount := 0
	startVideoId := ""
	endVideoId := ""

	for video := videos.Back(); video != nil; video = video.Prev() {
		videoObj := video.Value.(map[string]string)

		if isExistsVideo(videoObj["id"]) {
			continue
		}

		stmtIns, stmtInsErr := db.Prepare("INSERT INTO new_videos (id, title, post_datetime, status) VALUES( ?, ?, ?, ?)")
		if stmtInsErr != nil {
			panic(stmtInsErr.Error())
		}
		defer stmtIns.Close()

		// fmt.Println(videoObj["id"], " ", videoObj["datetime"], " ", videoObj["title"])
		insertCount++
		if startVideoId == "" {
			startVideoId = videoObj["id"]
		}
		endVideoId = videoObj["id"]

		_, insErr := stmtIns.Exec(videoObj["id"], videoObj["title"], videoObj["datetime"], 0)
		if insErr != nil {
			panic(insErr.Error())
		}
	}

	fmt.Println("datetime=[" + time.Now().String() + "] insertCount=[" + fmt.Sprint(insertCount) + "] startVideoId=[" + startVideoId + "] endVideoId=[" + endVideoId + "]")
}

func isExistsVideo(videoId string) bool {

	stmt, stmtErr := db.Prepare("SELECT count(id) FROM new_videos WHERE id = ?")
	if stmtErr != nil {
		panic(stmtErr.Error())
	}
	defer stmt.Close()

	var count string
	err := stmt.QueryRow(videoId).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		panic(err.Error())
	}

	if count == "1" {
		return true
	} else {
		return false
	}
}

func main() {
	loadConfig()
	db = getDbConnection()
	defer db.Close()

	lastVideoId, lastPostDateTime := selectLastCollectedVideo()
	tagsSlice := [][]string{
		[]string{"エンターテイメント", "音楽", "歌ってみた", "演奏してみた", "踊ってみた"},
		[]string{"動物", "料理", "自然", "旅行", "スポーツ", "ニコニコ動画講座", "車載動画", "歴史", "政治", "科学"},
		[]string{"ニコニコ技術部", "ニコニコ手芸部", "作ってみた", "描いてみた"},
		[]string{"例のアレ", "日記", "その他"},
		[]string{"VOCALOID", "ニコニコインディーズ"},
		[]string{"アニメ", "東方", "アイドルマスター", "ラジオ"},
		[]string{"ゲーム"}}

	videos := list.New()
	for _, tags := range tagsSlice {
		collectNewVideoByCategory(lastVideoId, lastPostDateTime, tags, videos)
	}

	if lastPostDateTime == "" {
		lastPostDateTime = "999999999999"
		for vi := videos.Front(); vi != nil; vi = vi.Next() {
			viMap := vi.Value.(map[string]string)
			if viMap["datetime"] < lastPostDateTime {
				lastPostDateTime = viMap["datetime"]
				lastVideoId = viMap["id"]
			}
		}
	}
	collectNewVideo(lastVideoId, lastPostDateTime, videos)
	registerNewVideos(videos)
}
