package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mmcdole/gofeed"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	bucketName    = "records"
	boltFileName  = "team-records.db"
	tgTokenEnv    = "TGBOT_TOKEN"
	tgChatEnv     = "TG_CHAT_ID"
	gitlabLinkEnv = "GITLAB_ATOM_LINK"
)

var duration = 5 * time.Second

type app struct {
	fp     *gofeed.Parser
	chat   *tb.Chat
	bucket string
	link   string
	db     *bolt.DB
	bot    *tb.Bot
}

func usage() {
	fmt.Printf(`Usage:
	%s=<telegram_bot_token> \
	%s=<telegram_chat_id> \
	%s=<link_to_gitlab_activity_atom> \
	./gitlab-atom-tgbot
`, tgTokenEnv, tgChatEnv, gitlabLinkEnv)
	os.Exit(1)
}

func main() {
	if (os.Getenv(tgTokenEnv) == "") || (os.Getenv(tgChatEnv) == "") || (os.Getenv(gitlabLinkEnv) == "") {
		usage()
	}
	a, err := newAtomTgBot()
	if err != nil {
		log.Fatal(err)
	}

	err = a.procActivity(true)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(duration)
	for range ticker.C {
		err := a.procActivity(false)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func newAtomTgBot() (*app, error) {
	db, err := bolt.Open(boltFileName, 0644, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("Create %s bucket %s failed", bucketName, boltFileName)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	b, err := tb.NewBot(tb.Settings{
		Token: os.Getenv(tgTokenEnv),
	})
	if err != nil {
		return nil, err
	}

	chatID, err := strconv.ParseInt(os.Getenv(tgChatEnv), 10, 64)
	if err != nil {
		return nil, err
	}

	return &app{
		fp: gofeed.NewParser(),
		chat: &tb.Chat{
			ID: chatID,
		},
		link:   os.Getenv(gitlabLinkEnv),
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

func (a *app) procActivity(initStart bool) error {
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

		if !initStart {
			_, err = a.bot.Send(
				a.chat,
				fmt.Sprintf("[%s](%s)", feed.Items[i].Title, feed.Items[i].Link),
				&tb.SendOptions{
					ParseMode:             tb.ModeMarkdown,
					DisableWebPagePreview: true,
				},
			)
			if err != nil {
				return err
			}
		}
		err = a.addRecord(feed.Items[i].GUID)
		if err != nil {
			return err
		}
	}

	return nil
}
