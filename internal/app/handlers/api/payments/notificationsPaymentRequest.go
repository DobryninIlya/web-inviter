package api_handler

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	h "main/internal/app/handlers/api"
	"main/internal/app/store/sqlstore"
	"main/internal/app/tg_api"
	"main/internal/payments"
	"net/http"
	"strconv"
	"strings"
)

func NewNotificationsPaymentRequestHandler(log *logrus.Logger, store sqlstore.StoreInterface, pay payments.Yokassa, tg *tg_api.APItg) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const path = "handlers.api.payments.NewCheckPaymentRequestHandler"
		var notification payments.Notification
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Log(logrus.ErrorLevel, path+": "+err.Error())
			h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
			return
		}
		json.Unmarshal(body, &notification)
		fmt.Println(notification)
		if notification.Event != "payment.succeeded" {
			h.RespondAPI(w, r, http.StatusOK, "not payment.succeeded")
			return
		}
		result, err := pay.CheckPaymentRequest(notification.Object.Id)
		if err != nil {
			log.Log(logrus.ErrorLevel, path+": "+err.Error())
			h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
			return
		}
		if result.Status == "succeeded" {
			transaction, err := store.API().GetTransaction(notification.Object.Id)
			channelID, err := store.API().GetChannelIDByType(transaction.Type)
			if err != nil {
				log.Log(logrus.ErrorLevel, path+": "+err.Error())
				h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
				return
			}
			link := tg.CreateInviteLink(log, channelID, 1)
			err = store.API().MakePremiumStatus(notification.Object.Id, link)
			if err != nil && err != sqlstore.ErrTransaсtionAlreadyEnded {
				log.Log(logrus.ErrorLevel, path+": "+err.Error()+" paymentID: "+notification.Object.Id)
				h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
				return
			}
			if err == sqlstore.ErrTransaсtionAlreadyEnded {
				tg.RevokeChatInviteLink(log, channelID, link)
				h.RespondAPI(w, r, http.StatusOK, "payment already ended")
				return
			}
			var tgClient int64
			clientIDS := strings.Trim(transaction.ClientID, "tg")
			if clientIDS != "" {
				tgClient, err = strconv.ParseInt(clientIDS, 10, 64)
				if err != nil {
					log.Log(logrus.ErrorLevel, path+": "+err.Error())
					h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
					return
				}
			}
			sended := tg.SendMessageTG(log, tgClient, "Вы успешно оплатили подписку. Ссылка на группу: "+link, "", 0)
			if !sended {
				log.Log(logrus.ErrorLevel, path+": Can't send message to client. Payment: "+notification.Object.Id)
				return
			}
		}
		h.RespondAPI(w, r, http.StatusOK, result.Status)
	}
}
