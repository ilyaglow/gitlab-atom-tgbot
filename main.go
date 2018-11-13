package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mmcdole/gofeed"
	tb "gopkg.in/tucnak/telebot.v2"
)

const bucketName = "records"

type app struct {
	fp     *gofeed.Parser
	chat   *tb.Chat
	bucket string
	link   string
	db     *bolt.DB
	bot    *tb.Bot
}

func main() {
	a, err := newRSSTgBot()
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		err := a.procUpdates()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func newRSSTgBot() (*app, error) {
	db, err := bolt.Open("team-records.db", 0644, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return errors.New("Create records bucket failed")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	b, err := tb.NewBot(tb.Settings{
		Token: os.Getenv("TGBOT_TOKEN"),
	})
	if err != nil {
		return nil, err
	}

	chatID, err := strconv.ParseInt(os.Getenv("TG_CHAT_ID"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &app{
		fp: gofeed.NewParser(),
		chat: &tb.Chat{
			ID: chatID,
		},
		link:   os.Getenv("GITLAB_RSS_LINK"),
		db:     db,
		bot:    b,
		bucket: bucketName,
	}, nil
}

func (a *app) recordExists(line string) (bool, error) {
	exists := false
	a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(a.bucket))
		v := b.Get([]byte(line))
		if v != nil {
			exists = true
		}
		return nil
	})
	return exists, nil
}

func (a *app) addRecord(line string) error {
	err := a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(a.bucket))
		err := b.Put([]byte(line), []byte(""))
		return err
	})

	return err
}

func (a *app) procUpdates() error {
	feed, err := a.fp.ParseURL(a.link)
	if err != nil {
		return err
	}

	for i := len(feed.Items) - 1; i >= 0; i-- {
		rex, err := a.recordExists(feed.Items[i].GUID)
		if err != nil {
			return err
		}
		if rex {
			continue
		}

		_, err = a.bot.Send(a.chat, fmt.Sprintf("[%s](%s)", feed.Items[i].Title, feed.Items[i].Link), &tb.SendOptions{
			ParseMode:             tb.ModeMarkdown,
			DisableWebPagePreview: true,
		})
		if err != nil {
			return err
		}
		err = a.addRecord(feed.Items[i].GUID)
		if err != nil {
			return err
		}
	}

	return nil
}
