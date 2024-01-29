package vk_app

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"log"
	"main/internal/app/handlers/api"
	pay_handler "main/internal/app/handlers/api/payments"
	"main/internal/app/store/sqlstore"
	"main/internal/app/tg_api"
	"main/internal/payments"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var secretKey = os.Getenv("SECRET_KEY")

type App struct {
	router *chi.Mux
	server *http.Server
	store  sqlstore.StoreInterface
	tgApi  *tg_api.APItg
	logger *logrus.Logger
	ctx    context.Context
	pay    payments.Yokassa
}

func newApp(ctx context.Context, store sqlstore.StoreInterface, bindAddr string, config Config) *App {
	router := chi.NewRouter()
	server := &http.Server{
		Addr:    bindAddr,
		Handler: router,
	}
	logger := logrus.New()
	a := &App{
		router: router,
		server: server,
		store:  store,
		logger: logger,
		tgApi:  tg_api.NewAPItg(config.BotToken),
		ctx:    ctx,
		pay:    payments.NewYokassa(config.ShopID, config.APIKey, logger),
	}
	a.configureRouter()
	return a
}

func (a *App) Close() error {
	err := a.server.Close()
	if err != nil {
		return err
	}
	return a.server.Close()
}

func (a *App) configureRouter() {
	a.router.Use(a.logRequest)
	//a.router.Use(a.APIMetricsMiddleware)
	//a.router.Use(imageStatusCodeHandler)
	a.router.Route("/payments", func(r chi.Router) {
		r.Use(a.parseURLParamsFromTelegramStart)
		r.Get("/request", pay_handler.NewPaymentRequestHandler(a.logger, a.store, a.pay))                              // Создание платежной заявки
		r.Get("/check/{payment_id}", pay_handler.NewCheckPaymentRequestHandler(a.logger, a.store, a.pay, a.tgApi))     // Ручная проверка статуса платежа
		r.Get("/done/{payment_id}", pay_handler.NewDonePaymentPageHandler(a.logger, a.store, a.pay, a.tgApi))          // Редирект после успешной оплаты,
		r.Post("/notifications", pay_handler.NewNotificationsPaymentRequestHandler(a.logger, a.store, a.pay, a.tgApi)) // Шлюз для автоматического получения вебхуков юкассы
		r.Get("/subscribe", pay_handler.NewMakePaymentPageHandler())                                                   // Стартовая страница с кнопокой подписки
	})
	a.router.Handle("/static/css/*", http.StripPrefix("/static/css/", cssHandler(http.FileServer(http.Dir(filepath.Join("internal", "app", "templates", "css"))))))
	a.router.Handle("/static/js/*", http.StripPrefix("/static/js/", http.FileServer(http.Dir(filepath.Join("internal", "app", "templates", "js")))))
	a.router.Handle("/static/img/*", http.StripPrefix("/static/img/", http.FileServer(http.Dir(filepath.Join("internal", "app", "templates", "img")))))
	a.router.Handle("/static/json/*", http.StripPrefix("/static/json/", http.FileServer(http.Dir(filepath.Join("internal", "app", "templates", "json")))))
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func cssHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		next.ServeHTTP(w, r)
	})
}

func (a *App) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := a.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}
		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
	})
}

func (a *App) parseURLParamsFromTelegramStart(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tgWebAppStartParam := r.URL.Query().Get("tgWebAppStartParam")
		if tgWebAppStartParam != "" {
			resultString, err := processQueryString(tgWebAppStartParam)
			if err != nil {
				log.Println(err)
			} else {
				updateRequestParams(r, resultString)
			}
		}
		rw := &responseWriter{w, http.StatusOK}
		h.ServeHTTP(rw, r)

		return
	})
}
func (a *App) checkSignTelegram(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := a.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
		})
		urlSign := r.FormValue("sign")
		if urlSign == "" {
			urlSign = r.URL.Query().Get("sign")
		}
		sign := api.GetSignForURLParams(r.URL.Query(), secretKey)
		if sign != urlSign {
			log.Log(
				logrus.WarnLevel,
				"the signature didn't match.",
			)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		rw := &responseWriter{w, http.StatusOK}
		h.ServeHTTP(rw, r)

		return
	})
}

func processQueryString(input string) (string, error) {
	// Декодируем URL-параметры
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		return "", err
	}

	// Заменяем тройной знак ___ на знак &
	processed := strings.ReplaceAll(decoded, "___", "&")
	processed = strings.ReplaceAll(processed, "---", "/")

	return processed, nil
}

func updateRequestParams(r *http.Request, queryString string) (*http.Request, error) {
	// Разбираем строку с URL-параметрами в url.Values
	queryValues, err := url.ParseQuery(queryString)
	if err != nil {
		return nil, err
	}

	// Обновляем параметры запроса в r.URL
	r.URL.RawQuery = queryValues.Encode()

	// Обновляем параметры запроса в r.Form
	r.Form = queryValues

	return r, nil
}

// Middleware для параллельной обработки хэндлеров
func ParallelHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			next.ServeHTTP(w, r)
		}()
		wg.Wait()
	})
}
