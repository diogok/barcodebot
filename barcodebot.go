package main

import (
	"bytes"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"io/ioutil"
	"kbot"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var botan string
var token string
var host string

var hredis string
var pool *redis.Pool

func handler(updt kbot.Update, outM chan<- kbot.OutMessage, _ chan<- kbot.OutQuery) {
	text := updt.Message.Text
	chatId := updt.Message.Chat.Id
	if text == "" {
	} else if text[0:1] == "/" {
		track(botan, text, updt.Message.From.Id)
		switch text {
		case "/start":
			Set(chatId, "lastCommand", "")
			outM <- kbot.OutMessage{Chat_id: updt.Message.Chat.Id, Text: "Choose a code to generate: /datamatrix, /qr, /code128, /code39, /ean, /2of5 or /codabar."}
		case "/datamatrix", "/qr", "/code128", "/ean", "/code39", "/2of5", "/codabar":
			Set(chatId, "lastCommand", text)
			outM <- kbot.OutMessage{Chat_id: updt.Message.Chat.Id, Text: "Send me the text to encode."}
		default:
			Set(chatId, "lastCommand", "")
			outM <- kbot.OutMessage{Chat_id: updt.Message.Chat.Id, Text: "Choose a code to generate: /datamatrix, /qr, /code128, /code39, /ean, /2of5 or /codabar."}
		}
	} else {
		last := Get(chatId, "lastCommand")
		if last == "" {
			outM <- kbot.OutMessage{Chat_id: updt.Message.Chat.Id, Text: "Choose a code to generate: /datamatrix, /qr, /code128, /code39, /ean, /2of5 or /codabar."}
		} else {
			Set(chatId, "lastCommand", "")

			var myUrl *url.URL
			myUrl, _ = url.Parse("https://barcode.kbots.net")
			myUrl.Path += last
			parameters := url.Values{}
			parameters.Add("content", text)
			parameters.Add("width", "256")
			myUrl.RawQuery = parameters.Encode()
			link := myUrl.String()
			outM <- kbot.OutMessage{Chat_id: chatId, Text: link}
		}
	}
}

func Set(id int, key string, value string) {
	conn := Conn()
	conn.Send("MULTI")
	conn.Send("SET", fmt.Sprintf("codebar:%d:%s", id, key), value)
	conn.Send("EXPIRE", fmt.Sprintf("codebar:%d:%s", id, key), TTL())
	conn.Do("EXEC")
}

func TTL() int {
	return 1 * 60 * 60
}

func Get(id int, key string) string {
	conn := Conn()
	value, _ := redis.String(conn.Do("GET", fmt.Sprintf("codebar:%d:%s", id, key)))
	return value
}

func Conn() redis.Conn {
	if pool != nil {
		return pool.Get()
	} else {
		pool = &redis.Pool{
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", fmt.Sprintf("%s:6379", hredis))
				if err != nil {
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
		return pool.Get()
	}
}

func track(token string, command string, user int) {
	resp, err := http.Post(fmt.Sprintf("https://api.botan.io/track?token=%s&name=%s&uid=%d", token, command, user), "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		log.Println("Error tracking", err)
	} else {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
}

func main() {
	botan = os.Getenv("BOTAN")
	token = os.Getenv("TOKEN")
	host = os.Getenv("HOST")
	hredis = os.Getenv("REDIS_HOST")
	bot := kbot.Bot{Token: token, Host: host, Handler: handler}
	_, done := kbot.Start(bot)
	<-done
}
