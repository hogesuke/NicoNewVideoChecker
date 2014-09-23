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
	videoIdRows, err := db.Query("SELECT id, post_datetime FROM new_videos ORDER BY post_datetime ASC, id ASC")
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

func registerVideoDetails(video Thumb, videoId string, postDatetime string) {
	insertVideo(video, videoId, postDatetime)
	registerTags(video.Thumb.Tags, videoId)
	registerContributor(video, videoId)
}

func insertVideo(video Thumb, videoId string, postDatetime string) {
	stmtIns, stmtErr := db.Prepare("INSERT INTO videos (id, title, description, thumbnail_url, post_datetime, length, view_counter, comment_counter, mylist_counter) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if stmtErr != nil {
		panic(stmtErr.Error())
	}
	defer stmtIns.Close()

	_, insErr := stmtIns.Exec(
		videoId,
		video.Thumb.Title,
		video.Thumb.Description,
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

func registerTags(tags []Tags, videoId string) {
	for i, _ := range tags {
		if tags[i].Domain != "jp" {
			continue
		}
		for j, _ := range tags[i].Tag {
			tagId := selectTagId(tags[i].Tag[j])
			if tagId == "" {
				insertTag(tags[i].Tag[j])
				tagId = selectTagId(tags[i].Tag[j])
			}
			insertVideoTagRelation(videoId, tagId)
		}
	}
}

func selectTagId(tag string) string {
	stmt, stmtErr := db.Prepare("SELECT id FROM tags WHERE tag = ?")
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

func insertTag(tag string) {
	stmtIns, stmtInsErr := db.Prepare("INSERT INTO tags (tag) VALUES(?)")
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

func insertVideoTagRelation(videoId string, tagId string) {
	stmtIns, stmtInsErr := db.Prepare("INSERT INTO videos_tags (video_id, tag_id) VALUES(?, ?)")
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

func registerContributor(video Thumb, videoId string) {
	exists := existsContributorId(video.Thumb.ContributorId)
	if !exists {
		insertContributor(video.Thumb.ContributorId, video.Thumb.ContributorName, video.Thumb.ContributorIconUrl)
	}
	insertVideoContributorRelation(videoId, video.Thumb.ContributorId)
}

func existsContributorId(contributorId string) bool {
	stmt, stmtErr := db.Prepare("SELECT id FROM contributors WHERE id = ?")
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

func insertContributor(contributorId string, contributorName string, contributorIconUrl string) {
	stmtIns, stmtInsErr := db.Prepare("INSERT INTO contributors (id, name, icon_url) VALUES(?, ?, ?)")
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

func insertVideoContributorRelation(videoId string, contributorId string) {
	stmtIns, stmtInsErr := db.Prepare("INSERT INTO videos_contributors (video_id, contributor_id) VALUES(?, ?)")
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

func main() {
	db = getDbConnection()
	defer db.Close()

	videoRows := selectNewVideos()
	for videoRows.Next() {
		var videoId string
		var postDatetime string
		videoRows.Scan(&videoId, &postDatetime)
		videoDetails := getVideoDetails(videoId)

		if videoDetails.Status != "fail" {
			registerVideoDetails(videoDetails, videoId, postDatetime)
		}
	}
}
