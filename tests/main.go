package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var IDs []string = []string{}

type post struct {
	Title    string    `json:"title"`
	BodyText string    `json:"bodytext"`
	Parent   string    `json:"parent"`
	TS       time.Time `json:"ts"`
}

type resbod struct {
	ID string `json:"replyID"`
	TS string `json:"timestamp"`
}

func main() {
	newPost()
}

func newPost() {
	for {
		sendJSON(mkJSON("root"), "submitForm")
		time.Sleep(2 * time.Second)
		sendJSON(mkJSON(IDs[rand.Intn(len(IDs))]), "submitForm")
		time.Sleep(2 * time.Second)
	}
}

func mkJSON(parentID string) []byte {
	b, err := json.Marshal(&post{
		BodyText: "genMessage()",
		Parent:   parentID,
	})
	if err != nil {
		log.Println(err)
	}
	return b
}

func sendJSON(bJSON []byte, rt string) {
	var url string = "https://walboard.xyz/" + rt
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bJSON))
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	rb := &resbod{}
	err = json.Unmarshal(body, rb)
	if err != nil {
		log.Println(err)
	}
	if len(rb.ID) > 6 {
		IDs = append(IDs, rb.ID)
	}
}

func genMessage() (message string) {
	symbols := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for j := 0; j <= rand.Intn(60)+15; j++ {
		var word string = ""
		for i := 0; i <= rand.Intn(15)+2; i++ {
			s := rand.Intn(len(symbols))
			word += symbols[s : s+1]
		}
		message += " " + word
	}
	return
}
