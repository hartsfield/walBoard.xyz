package main

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// home is displays the main page
func home(w http.ResponseWriter, r *http.Request) {
	var v viewData
	v.Order = "ranked"
	if len(postDBRank) < 20 {
		v.Stream = postDBRank[:]
	} else {
		v.Stream = postDBRank[:20]
	}
	exeTmpl(w, r, &v, "main.tmpl")
}
func pageInOrder(db []*post, r *http.Request, count int, v *viewData) map[string]string {
	var bb bytes.Buffer
	var nextCount string
	params, err := url.ParseQuery(strings.Split(r.RequestURI, "?")[1])
	if err != nil {
		log.Println(err)
	}
	if params["count"] == nil {
		params["count"] = append(params["count"], "0")
	}
	if params["count"][0] != "None" {
		count, err := strconv.Atoi(params["count"][0])
		if err != nil {
			log.Println(err)
		}
		if len(db) <= count+20 {
			v.Stream = db[count:]
			nextCount = "None"
		} else {
			v.Stream = db[count : count+20]
			nextCount = strconv.Itoa(count + 20)
		}
		err = templates.ExecuteTemplate(&bb, "stream.tmpl", v)
		if err != nil {
			log.Println(err)
		}
	}
	return map[string]string{
		"success":  "true",
		"template": bb.String(),
		"count":    nextCount,
	}
}

// getByChron returns 20 posts at a time in chronological order
func getByChron(w http.ResponseWriter, r *http.Request) {
	var count int = 20
	var v viewData
	v.Order = "chron"
	if len(strings.Split(r.RequestURI, "?")) > 1 {
		ajaxRes := pageInOrder(postDBChron, r, count, &v)
		ajaxResponse(w, ajaxRes)
	} else {
		v.Stream = postDBChron[count+(len(postDBChron)-count):]
		exeTmpl(w, r, &v, "main.tmpl")
	}
}

// getByRanked returns 20 posts at a time in ranked order.
func getByRanked(w http.ResponseWriter, r *http.Request) {
	var count int = 20
	var v viewData
	v.Order = "ranked"
	if len(strings.Split(r.RequestURI, "?")) > 1 {
		ajaxRes := pageInOrder(postDBRank, r, count, &v)
		ajaxResponse(w, ajaxRes)
	} else {
		if len(postDBRank) < count {
			v.Stream = postDBRank[:]
		} else {
			v.Stream = postDBRank[:count]
		}
		// v.Stream = postDBRank[count+(len(postDBRank)-count):]
		exeTmpl(w, r, &v, "main.tmpl")
	}
	// var v viewData
	// v.Order = "ranked"
	// var count int = 20
	// if len(strings.Split(r.RequestURI, "?")) > 1 {
	// 	params, err := url.ParseQuery(strings.Split(r.RequestURI, "?")[1])
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	if params["count"] == nil {
	// 		params["count"] = append(params["count"], "0")
	// 	}
	// 	if params["count"][0] != "None" {
	// 		count, err = strconv.Atoi(params["count"][0])
	// 		if err != nil {
	// 			log.Println(err)
	// 		}

	// 		var nextCount string
	// 		if len(postDBRank) < count {
	// 			v.Stream = postDBRank[count+(len(postDBRank)-count):]
	// 			nextCount = "None"
	// 		} else {
	// 			v.Stream = postDBRank[count+1 : count+20]
	// 			nextCount = strconv.Itoa(count + 20)
	// 		}
	// 		var bb bytes.Buffer
	// 		err = templates.ExecuteTemplate(&bb, "stream.tmpl", v)
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 		ajaxResponse(w, map[string]string{
	// 			"success":  "true",
	// 			"template": bb.String(),
	// 			"count":    nextCount,
	// 		})
	// 	}
	// } else {
	// 	if len(postDBRank) < count {
	// 		v.Stream = postDBRank[:]
	// 	} else {
	// 		v.Stream = postDBRank[:count]
	// 	}
	// 	exeTmpl(w, r, &v, "main.tmpl")
	// }
}

// viewPost returns a single post, with replies.
func viewPost(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.RequestURI, "/")
	var p post
	err := rdb.HGetAll(rdx, parts[len(parts)-1]).Scan(&p)
	if err != nil {
		log.Println(err)
	}
	if len(p.Id) == 11 {
		getAllChidren(&p, "RANK")
	} else {
		p.BodyText = "This post was automatically deleted."
	}
	var v viewData
	v.Stream = nil
	v.Stream = append(v.Stream, &p)
	v.ViewType = "post"
	exeTmpl(w, r, &v, "post.tmpl")
}

// handleForm verifies a users submissions and then adds it to the database.
func handleForm(w http.ResponseWriter, r *http.Request) {
	data, err := marshalPostData(r)
	if err != nil {
		log.Println(err)
	}
	parentExists, err := rdb.Exists(rdx, data.Parent).Result()
	if err != nil {
		log.Println(err)
	}

	log.Println(data, parentExists)
	if parentExists == 0 && data.Parent != "root" {
		ajaxResponse(w, map[string]string{
			"success":   "false",
			"replyID":   "",
			"timestamp": data.FTS,
		})
		return
	}
	if len(data.BodyText) < 5 || len(data.BodyText) > 1000 {
		ajaxResponse(w, map[string]string{"success": "false"})
		return
	}
	data.Id = genPostID(10)
	data.TS = time.Now()
	data.FTS = data.TS.Format("2006-01-02 03:04:05 pm")
	rdb.HSet(
		rdx, data.Id,
		"name", data.Author,
		"title", data.Title,
		"bodytext", data.BodyText,
		"id", data.Id,
		"ts", data.TS,
		"fts", data.FTS,
		"parent", data.Parent,
		"childCount", "0",
	)
	if data.Parent != "root" {
		rdb.ZAdd(rdx, data.Parent+":CHILDREN:CHRON", redis.Z{Score: float64(time.Now().UnixMilli()), Member: data.Id})
		rdb.ZAdd(rdx, data.Parent+":CHILDREN:RANK", redis.Z{Score: 0, Member: data.Id})
		bubbleUp(data)
	} else {
		rdb.ZAdd(rdx, "ANON:POSTS:CHRON", redis.Z{Score: float64(time.Now().UnixMilli()), Member: data.Id})
		rdb.ZAdd(rdx, "ANON:POSTS:RANK", redis.Z{Score: 0, Member: data.Id})
		// popLast()
	}
	ajaxResponse(w, map[string]string{
		"success":   "true",
		"replyID":   data.Id,
		"timestamp": data.FTS,
	})
	beginCache()
}
