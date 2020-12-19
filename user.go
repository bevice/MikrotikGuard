package main

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"os"
)

type User struct {
	Name   string `json:"name"`
	ChatID int64  `json:"chat_id"`
}

type Users struct {
	Users []*User `json:"users"`
}

func NewUsersFromJSON(fileName string) *Users {
	var users Users
	f, err := os.Open(fileName)
	if err != nil {
		// Файла нет или ошибка
		log.Printf("Ошибка открытия файла %s: %v", fileName, err)
		return &users
	}
	defer func() {
		_ = f.Close()
	}()
	byteData, err := ioutil.ReadAll(f)
	if err != nil {
		// Ошибка чтения
		log.Printf("Ошибка чтения файла %s: %v", fileName, err)
		return &users
	}

	err = json.Unmarshal(byteData, &users)
	if err != nil {
		// Ошибка разбора
		log.Printf("Ошибка разбора файла %s: %v", fileName, err)
	}

	return &users
}
func (users *Users) SaveJSON(fileName string) error {
	var err error
	data, err := json.MarshalIndent(users, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, data, 0644)
	return err
}

func (users *Users) IsAuthorised(chatId int64) bool {
	for _, user := range users.Users {
		if user.ChatID == chatId {
			return true
		}
	}
	return false
}
func (users *Users) AddUser(chatId int64, userName string) {
	user := User{
		Name:   userName,
		ChatID: chatId,
	}
	users.Users = append(users.Users, &user)
}
func (user *User) Send(text string) {
	var (
		msg tgbotapi.Chattable
	)

	tmp := tgbotapi.NewMessage(user.ChatID, text)
	tmp.ParseMode = "markdown"
	msg = tmp

	_, err := TheBot.Send(msg)

	if err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}
}

func (users *Users) Send(text string) {
	for _, user := range users.Users {

		user.Send(text)

	}
}

func (users *Users) GetUserByChatID(chatId int64) *User {
	for _, user := range users.Users {
		if user.ChatID == chatId {
			return user
		}
	}
	return nil
}