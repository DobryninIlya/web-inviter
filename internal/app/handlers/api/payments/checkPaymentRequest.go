package api_handler

import (
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	h "main/internal/app/handlers/api"
	"main/internal/app/store/sqlstore"
	"main/internal/app/tg_api"
	"main/internal/app/tools"
	"main/internal/payments"
	"net/http"
	"strconv"
	"strings"
)

func NewCheckPaymentRequestHandler(log *logrus.Logger, store sqlstore.StoreInterface, pay payments.Yokassa, tg *tg_api.APItg) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const path = "handlers.api.payments.NewCheckPaymentRequestHandler"
		uid := chi.URLParam(r, "payment_id")
		transaction, err := store.API().GetPaymentRequestByUID(uid)
		result, err := pay.CheckPaymentRequest(uid)
		extID := strings.TrimSpace(transaction.ExtID)
		if err != nil {
			log.Log(logrus.ErrorLevel, path+": "+err.Error())
			h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
			return
		}
		if result.Status == "succeeded" {
			channelID, err := store.API().GetChannelIDByType(transaction.Type)
			if err != nil {
				log.Log(logrus.ErrorLevel, path+": "+err.Error())
				h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
				return
			}
			link := tg.CreateInviteLink(log, channelID, 1)
			err = store.API().MakePremiumStatus(extID, link)
			if err != nil && err != sqlstore.ErrTransaсtionAlreadyEnded {
				log.Log(logrus.ErrorLevel, path+": "+err.Error()+" paymentID: "+extID)
				h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
				return
			}
			if err == sqlstore.ErrTransaсtionAlreadyEnded {
				tg.RevokeChatInviteLink(log, channelID, link)
				page, err := tools.GetDonePaymentTemplate()
				if err != nil {
					log.Log(logrus.ErrorLevel, path+": "+err.Error())
					h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(page))
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
				log.Log(logrus.ErrorLevel, path+": Can't send message to client. Payment: "+extID)
				return
			}
		}
		h.RespondAPI(w, r, http.StatusOK, result.Status)
	}
}
