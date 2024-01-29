package tg_api

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	model "main/internal/app/model"
	"net/http"
	"net/url"
)

const tgTemplate = "https://api.telegram.org/bot%s/%s%s"
const tgSendMethod = "sendMessage?"

type APItg struct {
	tgToken    string
	tgTemplate string
}

func NewAPItg(token string) *APItg {
	return &APItg{
		tgToken:    token,
		tgTemplate: tgTemplate,
	}
}

func (s APItg) SendMessageTG(log *logrus.Logger, uId int64, message string, buttons string, threadId int) bool {
	if uId == 0 {
		log.Printf("Попытка отправить сообщение некорректному айди")
		return false
	}
	params := fmt.Sprintf("chat_id=%v&text=%v&reply_markup=%s&parse_mode=markdown&message_thread_id=%v",
		uId,
		url.QueryEscape(message),
		url.QueryEscape(buttons),
		threadId,
	)
	url := fmt.Sprintf(s.tgTemplate, s.tgToken, tgSendMethod, params)
	resp, err := http.Get(url)
	if resp.StatusCode != 200 {
		log.Printf("Ошибка отправки сообщения в телеграм. Статус код: %v", resp.StatusCode)
		return false
	}
	if err != nil {
		log.Printf("Ошибка API. Отправка сообщений: %v", err)
		return false
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("TG: При этом возникла ошибка: %v", err)
		return false
	}
	defer resp.Body.Close()
	return true
}

func (s APItg) CreateInviteLink(log *logrus.Logger, chatID int64, memberLimit int) string {
	params := fmt.Sprintf("chat_id=%v&member_limit=%v", chatID, memberLimit)
	url := fmt.Sprintf(s.tgTemplate, s.tgToken, "createChatInviteLink?", params)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("Ошибка получения ссылки приглашения. Статус код: %v", resp.StatusCode)
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка API. Получение ссылки приглашения: %v", err)
		return ""
	}
	defer resp.Body.Close()
	var inviteLinkResponse model.InviteChatLinkResponse
	err = json.Unmarshal(body, &inviteLinkResponse)
	if err != nil {
		log.Printf("Ошибка API. Получение ссылки приглашения: %v", err)
		return ""
	}
	return inviteLinkResponse.Result.InviteLink
}

func (s APItg) RevokeChatInviteLink(log *logrus.Logger, chatID int64, inviteLink string) bool {
	params := fmt.Sprintf("chat_id=%v&invite_link=%v", chatID, inviteLink)
	url := fmt.Sprintf(s.tgTemplate, s.tgToken, "revokeChatInviteLink?", params)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("Ошибка отзыва ссылки приглашения. Статус код: %v", resp.StatusCode)
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка API. Отзыв ссылки приглашения: %v", err)
		return false
	}
	defer resp.Body.Close()
	var revokeLinkResponse model.RevokeLinkResponse
	err = json.Unmarshal(body, &revokeLinkResponse)
	if err != nil {
		log.Printf("Ошибка API. Отзыв ссылки приглашения: %v", err)
		return false
	}
	return revokeLinkResponse.Ok
}
