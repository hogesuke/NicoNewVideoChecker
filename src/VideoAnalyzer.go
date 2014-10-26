package main

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
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
	videoIdRows, err := db.Query("SELECT id, post_datetime FROM new_videos WHERE status = 0 ORDER BY post_datetime ASC, id ASC")
	if err != nil && err != sql.ErrNoRows {
		panic(err.Error())
	}

	return videoIdRows
}

func getVideoDetails(videoId string) Thumb {
	url := "http://ext.nicovideo.jp/api/getthumbinfo/sm"
	res, err := http.Get(url + videoId)
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err.Error())
	}

	videoXml := Thumb{Thumb: Video{
		Title: "",
		Description: "",
		ThumbnailUrl: "",
		Length: "",
		ViewCounter: "",
		CommentNum: "",
		MylistCounter: "",
		Tags: nil,
		ContributorId: "",
		ContributorName: "",
		ContributorIconUrl: ""},
		Status: ""}
	parseErr := xml.Unmarshal([]byte(body), &videoXml)
	if parseErr != nil {
		panic(parseErr.Error())
	}

	return videoXml
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
	ContributorId string `xml:"user_id"`
	ContributorName string `xml:"user_nickname"`
	ContributorIconUrl string `xml:"user_icon_url"`
}

type Tags struct {
	Domain string `xml:"domain,attr"`
	Tag []string `xml:"tag"`
}

func registerVideoDetails(tx *sql.Tx, video Thumb, videoId string, postDatetime string) {
	insertVideo(tx, video, videoId, postDatetime)
	registerTags(tx, video.Thumb.Tags, videoId)
	registerContributor(tx, video, videoId)
	updateNewVideo(tx, videoId, 1)
}

func insertVideo(tx *sql.Tx, video Thumb, videoId string, postDatetime string) {
	stmtIns, stmtErr := tx.Prepare(`
	INSERT INTO videos
	(id, title, description, contributor_id, contributor_name,
	thumbnail_url, post_datetime, length,
	view_counter, comment_counter, mylist_counter)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if stmtErr != nil {
		panic(stmtErr.Error())
	}
	defer stmtIns.Close()

	_, insErr := stmtIns.Exec(
		videoId,
		video.Thumb.Title,
		video.Thumb.Description,
		video.Thumb.ContributorId,
		video.Thumb.ContributorName,
		video.Thumb.ThumbnailUrl,
		postDatetime,
		video.Thumb.Length,
		video.Thumb.ViewCounter,
		video.Thumb.CommentNum,
		video.Thumb.MylistCounter)

	if insErr != nil {
		panic(insErr.Error())
	}
}

func registerTags(tx *sql.Tx, tags []Tags, videoId string) {
	for i, _ := range tags {
		if tags[i].Domain != "jp" {
			continue
		}
		for j, _ := range tags[i].Tag {
			tagId := selectTagId(tx, tags[i].Tag[j])
			if tagId == "" {
				insertTag(tx, tags[i].Tag[j])
				tagId = selectTagId(tx, tags[i].Tag[j])
			}
			insertVideoTagRelation(tx, videoId, tagId)
		}
	}
}

func selectTagId(tx *sql.Tx, tag string) string {
	stmt, stmtErr := tx.Prepare("SELECT id FROM tags WHERE tag = ?")
	if stmtErr != nil {
		panic(stmtErr.Error())
	}
	defer stmt.Close()

	var tagId string
	err := stmt.QueryRow(tag).Scan(&tagId)
	if err != nil && err != sql.ErrNoRows {
		panic(err.Error())
	}

	return tagId
}

func insertTag(tx *sql.Tx, tag string) {
	stmtIns, stmtInsErr := tx.Prepare("INSERT INTO tags (tag) VALUES(?)")
	if stmtInsErr != nil {
		panic(stmtInsErr.Error())
	}
	defer stmtIns.Close()

	_, insErr := stmtIns.Exec(tag)
	if insErr != nil {
		panic(insErr.Error())
	}
	defer stmtIns.Close()
}

func insertVideoTagRelation(tx *sql.Tx, videoId string, tagId string) {
	stmtIns, stmtInsErr := tx.Prepare("INSERT INTO videos_tags (video_id, tag_id) VALUES(?, ?)")
	if stmtInsErr != nil {
		panic(stmtInsErr.Error())
	}
	defer stmtIns.Close()

	_, insErr := stmtIns.Exec(videoId, tagId)
	if insErr != nil {
		panic(insErr.Error())
	}
	defer stmtIns.Close()
}

func registerContributor(tx *sql.Tx, video Thumb, videoId string) {
	exists := existsContributorId(tx, video.Thumb.ContributorId)
	if !exists {
		insertContributor(tx, video.Thumb.ContributorId, video.Thumb.ContributorName, video.Thumb.ContributorIconUrl)
	}
	insertVideoContributorRelation(tx, videoId, video.Thumb.ContributorId)
}

func existsContributorId(tx *sql.Tx, contributorId string) bool {
	stmt, stmtErr := tx.Prepare("SELECT id FROM contributors WHERE id = ?")
	if stmtErr != nil {
		panic(stmtErr.Error())
	}
	defer stmt.Close()

	var selectId string
	err := stmt.QueryRow(contributorId).Scan(&selectId)
	if err != nil && err != sql.ErrNoRows {
		panic(err.Error())
	}

	if selectId == "" {
		return false
	}
	return true
}

func insertContributor(tx *sql.Tx, contributorId string, contributorName string, contributorIconUrl string) {
	stmtIns, stmtInsErr := tx.Prepare("INSERT INTO contributors (id, name, icon_url) VALUES(?, ?, ?)")
	if stmtInsErr != nil {
		panic(stmtInsErr.Error())
	}
	defer stmtIns.Close()

	_, insErr := stmtIns.Exec(contributorId, contributorName, contributorIconUrl)
	if insErr != nil {
		panic(insErr.Error())
	}
	defer stmtIns.Close()
}

func insertVideoContributorRelation(tx *sql.Tx, videoId string, contributorId string) {
	stmtIns, stmtInsErr := tx.Prepare("INSERT INTO videos_contributors (video_id, contributor_id) VALUES(?, ?)")
	if stmtInsErr != nil {
		panic(stmtInsErr.Error())
	}
	defer stmtIns.Close()

	_, insErr := stmtIns.Exec(videoId, contributorId)
	if insErr != nil {
		panic(insErr.Error())
	}
	defer stmtIns.Close()
}

func updateNewVideo(tx *sql.Tx, video_id string, status int) {
	stmtUpd, stmtUpdErr := tx.Prepare("UPDATE new_videos SET status = ? WHERE id = ?")
	if stmtUpdErr != nil {
		panic(stmtUpdErr.Error())
	}
	defer stmtUpd.Close()

	_, updErr := stmtUpd.Exec(status, video_id)
	if updErr != nil {
		panic(updErr.Error())
	}
	defer stmtUpd.Close()
}

func main() {
	db = getDbConnection()
	defer db.Close()

	videoRows := selectNewVideos()
	for videoRows.Next() {
		var videoId string
		var postDatetime string
		videoRows.Scan(&videoId, &postDatetime)
		videoDetails := getVideoDetails(videoId)

		tx, err := db.Begin()
		if err != nil {
			panic(err.Error())
		}
		defer func() {
			if recoveredErr := recover(); recoveredErr !=nil {
				tx.Rollback()
				panic(recoveredErr)
			}
		}()

		if videoDetails.Status == "fail" {
			updateNewVideo(tx, videoId, 9)
		} else {
			registerVideoDetails(tx, videoDetails, videoId, postDatetime)
		}
		tx.Commit()
	}
}
