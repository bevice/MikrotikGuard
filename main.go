package main


/**
	Environment variables:
    TG_TOKEN - Telegram bot token (use @BotFather to get it)
	TG_PASSWORD - Say this word to authorize
	LOGGER_BIND - address and port for bind Syslog server, ex: "0.0.0.0:514"
	DATA_DIR 	- folder, that contains users.json file, RW permissions needed
 */

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mcuadros/go-syslog.v2"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)
/* default phrases for search in log messages, if message isn't contents it, it will be ignored" */
var phrases = []string{
	"logged in",
	"logged out",
	"failure for user",
	"changed by",
}



const (
	USERS_FILENAME = "users.json"
	PHRASES_FILENAME = "phrases.json"
)

var (
	TheBot *tgbotapi.BotAPI
	users  *Users
)


func checkCritical(err error) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		_, fn = path.Split(fn) // Нас интересует только имя файла
		log.Panicf("Critical [%s:%d]: %v", fn, line, err)
	}

}

func showError(err error, text string) bool {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		_, fn = path.Split(fn) // Нас интересует только имя файла
		log.Printf("Error[%s:%d]: %s [%v]", fn, line, text, err)
		return true
	}
	return false
}



func filter(msg string) bool {
	for _, s := range phrases {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func botUpdates(bot *tgbotapi.BotAPI, msgs chan *tgbotapi.Message) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		msgs <- update.Message
	}
}

func requestAuth(msg *tgbotapi.Message, TG_PASSWORD string) {
	if msg.Text != TG_PASSWORD {
		r := tgbotapi.NewMessage(msg.Chat.ID, "Say friend and Enter!")
		_, _ = TheBot.Send(r)
	} else {

		users.AddUser(msg.Chat.ID, msg.Chat.UserName)
		if err := users.SaveJSON(getFileName(USERS_FILENAME)); err != nil {

		}
		r := tgbotapi.NewMessage(msg.Chat.ID, "Welcome")
		_, _ = TheBot.Send(r)
	}
}

func getFileName(fileName string) string {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir != "" {
		return path.Join(dataDir, fileName)
	} else {
		return path.Join("data", fileName)
	}

}
func reply(msg *tgbotapi.Message) {
	if user := users.GetUserByChatID(msg.Chat.ID); user != nil {
		user.Send("All is ok, wait for message!")
	}
}
func main() {
	var err error

	var TG_TOKEN = os.Getenv("TG_TOKEN")
	var TG_PASSWORD = os.Getenv("TG_PASSWORD")
	var LOGGER_BIND = os.Getenv("LOGGER_BIND")

	users = NewUsersFromJSON(getFileName("users.json"))

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	msgchan := make(chan *tgbotapi.Message)



	TheBot, err = tgbotapi.NewBotAPI(TG_TOKEN)
	go botUpdates(TheBot, msgchan)
	checkCritical(err)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)
	err = server.ListenUDP(LOGGER_BIND)
	checkCritical(err)
	err = server.Boot()
	checkCritical(err)

	go server.Wait()
	for {
		select {
		case msg := <-msgchan:
			if !users.IsAuthorised(msg.Chat.ID) {
				go requestAuth(msg, TG_PASSWORD)
			} else {
				reply(msg)
			}
		case logParts := <-channel:
			logMessage := fmt.Sprintf("%v", logParts["content"])
			logHost := fmt.Sprintf("%v", logParts["hostname"])
			logTag := fmt.Sprintf("%v", logParts["tag"])
			logSrc := strings.Split(fmt.Sprintf("%v", logParts["client"]), ":")[0]
			if filter(logMessage) {
				messageText := fmt.Sprintf("Host: _%s_ (%s)\n```\n%s: %s```", logHost, logSrc, logTag, logMessage)
				users.Send(messageText)
			}
		}

	}

}
