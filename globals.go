package main

import (
	"context"
	"html/template"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	appConf     *config            = readConf()
	servicePort string             = ":" + appConf.App.Port
	logFilePath string             = appConf.App.Env["logFilePath"]
	templates   *template.Template = template.New("")
	AppName     string             = appConf.App.Name
	rdb                            = redis.NewClient(&redis.Options{
		Addr:     ":6379",
		Password: "",
		DB:       6,
	})
	// redis context
	rdx = context.Background()

	// Database caches
	postDBChron []*post
	postDBRank  []*post
)

type env map[string]string

type config struct {
	App    app    `json:"app"`
	GCloud gcloud `json:"gcloud"`
}

type app struct {
	Name       string `json:"name"`
	DomainName string `json:"domain_name"`
	Version    string `json:"version"`
	Env        env    `json:"env"`
	Port       string `json:"port"`
	AlertsOn   bool   `json:"alertsOn"`
	TLSEnabled bool   `json:"tls_enabled"`
	Repo       string `json:"repo"`
}

type gcloud struct {
	Command   string `json:"command"`
	Zone      string `json:"zone"`
	Project   string `json:"project"`
	User      string `json:"user"`
	LiveDir   string `json:"livedir"`
	ProxyConf string `json:"proxyConf"`
}

// post is the structure of a user post. Posts are created by users and stored
// in redis.
type post struct {
	Title  string `json:"title" redis:"title"`
	Id     string `json:"id" redis:"id"`
	Author string `json:"author,name" redis:"author"`
	// timestamp
	TS time.Time `json:"ts" redis:"ts"`
	// formatted time stamp
	FTS      string `json:"fts" redis:"fts"`
	BodyText string `json:"bodytext" redis:"bodytext"`
	// TODO: implment nonce
	Nonce      string  `json:"nonce" redis:"nonce"`
	Children   []*post `json:"children" redis:"children"`
	ChildCount int     `json:"childCount" redis:"childCount"`
	Parent     string  `json:"parent" redis:"parent"`
	// used for pagification
	PostCount string `json:"postCount" redis:"postCount"`
}
