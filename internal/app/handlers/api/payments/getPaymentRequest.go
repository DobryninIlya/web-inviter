package api_handler

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	h "main/internal/app/handlers/api"
	"main/internal/app/model"
	"main/internal/app/store/sqlstore"
	"main/internal/app/tools"
	"main/internal/payments"
	"net/http"
	"time"
)

const (
	PayGateway = "https://yoomoney.ru/checkout/payments/v2/contract?orderId="
)

func NewPaymentRequestHandler(log *logrus.Logger, store sqlstore.StoreInterface, pay payments.Yokassa) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const path = "handlers.api.payments.NewPaymentRequestHandler"
		//subscribeLevel := chi.URLParam(r, "level") // month, half-year, year
		url := r.URL.Query()
		subscribeLevel := url.Get("channel") // month, half-year, year
		var amount string
		switch subscribeLevel {
		case "test":
			amount = "150"
		case "test2":
			amount = "700"
		case "poop_zemli":
			amount = "1100"
		default:
			log.Log(logrus.ErrorLevel, path+": wrong subscribe channel, setted amount 1000")
			amount = "1000"
			//h.ErrorHandlerAPI(w, r, http.StatusBadRequest, errors.New("wrong subscribe level"))
			//return
		}
		clientID := url.Get("client_id")
		if clientID == "" {
			log.Log(logrus.ErrorLevel, path+": empty client id")
			h.ErrorHandlerAPI(w, r, http.StatusBadRequest, errors.New("empty client id"))
			return
		}
		idempotenceKey := fmt.Sprintf("%v:%v", time.Now().Unix(), clientID)
		uid := tools.RandStringBytes(32)
		payment, err := pay.PaymentRequest(payments.YokassaPayment{
			Amount: payments.Amount{
				Value:    amount,
				Currency: "RUB",
			},
			Capture: true,
			Confirmation: payments.Confirmation{
				Type:      "redirect",
				ReturnUrl: "https://literaturaforheart.ru/payments/done/" + uid, //TODO поменять return_URL
			},
			Description: "Премиум подписка Бот расписание КАИ",
		}, idempotenceKey)
		if err != nil {
			log.Log(logrus.ErrorLevel, path+": "+err.Error())
			h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
			return
		}
		err = store.API().SaveTransaction(model.Transaction{
			UID:      uid,
			ExtID:    payment.Id,
			ClientID: clientID,
			Type:     subscribeLevel,
		})
		if err != nil {
			log.Log(logrus.ErrorLevel, path+": "+err.Error())
			h.ErrorHandlerAPI(w, r, http.StatusInternalServerError, err)
			return
		}
		h.RespondAPI(w, r, http.StatusOK, PayGateway+payment.Id)
	}
}
