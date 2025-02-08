package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	gt "github.com/bas24/googletranslatefree"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	api    = "hugging face token"
	urlImg = "https://api-inference.huggingface.co/models/stabilityai/stable-diffusion-xl-base-1.0"
)

var img bool = false

func main() {
	bot, err := tg.NewBotAPI("tg bot token")
	if err != nil {
		fmt.Println("Проблема с TG API: ", err)
		return
	}

	u := tg.NewUpdate(60)
	u.Timeout = 0
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		keyboard := tg.NewReplyKeyboard(
			tg.NewKeyboardButtonRow(
				tg.NewKeyboardButton("сгенерировать картинку"),
			),
		)

		if update.Message.Text == "/start" {
			msg := tg.NewMessage(update.Message.Chat.ID, "Напиши промпт для генерации картинок!")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}

		if update.Message.Text == "сгенерировать картинку" {
			msg := tg.NewMessage(update.Message.Chat.ID, "Напишите промпт для генерирования картинки")
			msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
			bot.Send(msg)
			img = true
		}

		if update.Message.Text != "" && img == true && update.Message.Text != "сгенерировать картинку" {
			msg := tg.NewMessage(update.Message.Chat.ID, "Подождите")
			wait, _ := bot.Send(msg)
			result, _ := gt.Translate(update.Message.Text, "ru", "en")
			fmt.Println(result)

			path := generateImg(result)
			if path == "" {
				delete := tg.NewDeleteMessage(update.Message.Chat.ID, wait.MessageID)
				bot.Request(delete)
				msg := tg.NewMessage(update.Message.Chat.ID, "Не получилось сделать картинку. Попробуйте ещё раз!")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
				img = false
				break
			}

			file := tg.NewPhoto(update.Message.Chat.ID, tg.FilePath(path))
			delete := tg.NewDeleteMessage(update.Message.Chat.ID, wait.MessageID)
			bot.Request(delete)
			file.ReplyMarkup = keyboard
			bot.Send(file)
			img = false
			os.Remove(path)
			if err != nil {
				fmt.Println("Ошибка при открытии файла:", err)
				return
			}
		}
	}
}

func generateImg(prompt string) string {
	payload := map[string]interface{}{
		"inputs": prompt,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Ошибка в кодировании: ", err)
		return ""
	}

	req, err := http.NewRequest("POST", urlImg, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return ""
	}

	req.Header.Set("Authorization", "Bearer "+api)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "image/png")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка в запросе: ", err)
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Println("Ошибка API:", resp.Status, string(bodyBytes))
		return ""
	}

	imagedata, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Не удалось прочитать: ", err)
		return ""
	}

	file, _ := os.Create("gen.png")
	err = os.WriteFile("gen.png", imagedata, 0644)
	if err != nil {
		fmt.Println("Ошибка при записи: ", err)
		return ""
	}

	file.Close()
	return "gen.png"
}
