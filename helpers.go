package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// lastCached is how many milliseconds ago the database was last cached
var lastCached time.Time

// beginCache will cache the database no more than every 3 seconds. This
// function is run on startup, and when a post or reply is added to the
// database. In the event that many posts are being posted at once, the
// function is designed to only rebuild the cache every 3 seconds. This could
// be adjusted if needed.
func beginCache() {
	if time.Since(lastCached).Milliseconds() > 500 {
		lastCached = time.Now()
		// Race condition(?) prevention. Say we have two users posting
		// consecutively. User_1 submits a post and this triggers a
		// rebuild of the cache. User_2 submits a post 1 second later.
		// If we didn't have this delay, the User_2's post would not
		// get cached, because the rebuild triggered by User_1 would
		// have already started.
		// By delaying the rebuild for 3 seconds we ensure all posts
		// are cached, even those that don't trigger a re-cache
		// automatically.
		// This function is non-blocking (is executed in a goroutine)
		time.AfterFunc(500*time.Millisecond, func() { buildDB() })
	}
}

// buildDB is used to cache the redis database
func buildDB() {
	postDBChron = nil
	postDBRank = nil
	var ids []string
	opts := &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  -1,
	}
	err := rdb.ZRevRangeByScore(rdx, "ANON:POSTS:CHRON", opts).ScanSlice(&ids)
	if err != nil {
		log.Println(err)
	}
	for _, id := range ids {
		var p post
		rdb.HGetAll(rdx, id).Scan(&p)
		getAllChidren(&p, "CHRON")
		postDBChron = append(postDBChron, &p)
	}

	var rankedIds []string
	err = rdb.ZRevRangeByScore(rdx, "ANON:POSTS:RANK", opts).ScanSlice(&rankedIds)
	if err != nil {
		log.Println(err)
	}
	for _, id := range rankedIds {
		var p post
		rdb.HGetAll(rdx, id).Scan(&p)
		getAllChidren(&p, "RANK")
		postDBRank = append(postDBRank, &p)
	}
}

// getAllChidren is a recursive function used to get all the children of a post
func getAllChidren(po *post, suffix string) {
	var ids []string
	opts := &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  -1,
	}
	err := rdb.ZRevRangeByScore(rdx, po.Id+":CHILDREN:"+suffix, opts).ScanSlice(&ids)
	if err != nil {
		log.Println(err)
	}

	for _, id := range ids {
		var p post
		rdb.HGetAll(rdx, id).Scan(&p)
		getAllChidren(&p, suffix)
		po.Children = append(po.Children, &p)
	}
}

// bubbleUp is a recursive function used to increment the ranks of posts as
// they're replied to.
func bubbleUp(p *post) {
	str, err := rdb.HGet(rdx, p.Id, "childCount").Result()
	if err != nil {
		log.Println(err)
	}

	if len(str) > 0 {
		num, err := strconv.Atoi(str)
		if err != nil {
			log.Println(err)
		}

		rdb.HSet(rdx, p.Id, "childCount", fmt.Sprint(num+1))
		if p.Parent == "root" {
			rdb.ZIncrBy(rdx, "ANON:POSTS:RANK", 1, p.Id)
		}

		if p.Parent != "root" {
			rdb.ZIncrBy(rdx, p.Parent+":CHILDREN:RANK", 1, p.Id)
			var po post
			err = rdb.HGetAll(rdx, p.Parent).Scan(&po)
			if err != nil {
				log.Println(err)
			}

			bubbleUp(&po)
		}
	}
}

// popLast removes the oldest post in the database, and is used to maintain a
// count of 99 active threads
func popLast() {
	length_, err := rdb.ZCount(rdx, "ANON:POSTS:CHRON", "-inf", "+inf").Result()
	if err != nil {
		log.Println(err)
	}

	if length_ >= 100 {
		lastPostID, err := rdb.ZRange(rdx, "ANON:POSTS:CHRON", 0, 0).Result()
		if err != nil {
			log.Println(err)
		}

		rdb.ZRemRangeByRank(rdx, "ANON:POSTS:CHRON", 0, 0)
		rdb.ZRem(rdx, "ANON:POSTS:RANK", lastPostID)
		rdb.Del(rdx, lastPostID[0]+":CHILDREN:RANK")
		rdb.Del(rdx, lastPostID[0]+":CHILDREN:CHRON")
		rdb.Del(rdx, lastPostID...)
		// beginCache()
	}
}

// exeTmpl is used to build and execute an html template.
func exeTmpl(w http.ResponseWriter, r *http.Request, view *viewData, tmpl string) {
	if view == nil {
		view = &viewData{}
	}
	view.AppName = AppName
	err := templates.ExecuteTemplate(w, tmpl, view)
	if err != nil {
		log.Println(err)
	}
}

// ajaxResponse is used to respond to ajax requests with arbitrary data in the
// format of map[string]string
func ajaxResponse(w http.ResponseWriter, res map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(err)
	}
}

// genPostID generates a post ID
func genPostID(length int) (ID string) {
	symbols := "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i <= length; i++ {
		s := rand.Intn(len(symbols))
		ID += symbols[s : s+1]
	}
	return
}

// marshalPostData is used to marshal a JSON string into a *post struct
func marshalPostData(r *http.Request) (*post, error) {
	t := &post{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(t)
	if err != nil {
		return t, err
	}
	return t, nil
}

func readConf() *config {
	b, err := os.ReadFile("./bolt.conf.json")
	if err != nil {
		log.Println(err)
	}
	c := config{}
	err = json.Unmarshal(b, &c)
	if err != nil {
		log.Println(err)
	}
	return &c
}

// setupLogging sets output flags and the file for logging
func setupLogging() (f *os.File) {
	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)

	return
}
