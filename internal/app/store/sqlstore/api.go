package sqlstore

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"main/internal/app/model"
	"strings"
	"time"
)

var (
	ErrBadNews                 = errors.New("новость не подходит для публикации")
	ErrBadPhoto                = errors.New("новость не подходит для публикации, не содержит фото")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserNotFoundInFirebase  = errors.New("user not found in firebase")
	ErrBadMobileUserInfo       = errors.New("incorrect mobile user info")
	ErrAlreadyRegistered       = errors.New("user already registered")
	ErrTransaсtionAlreadyEnded = errors.New("transaction already ended")
)

type API interface {
}

// ApiRepository реализует работу API с хранилищем базы данных
type ApiRepository struct {
	store            *Store
	ConfirmationCode string
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=ApiRepositoryInterface
type ApiRepositoryInterface interface {
	SaveTransaction(transaction model.Transaction) error
	MakePremiumStatus(uid string, inviteLink string) error
	GetTransaction(uid string) (model.Transaction, error)
	GetPaymentRequestByUID(uid string) (model.Transaction, error)
	CheckPremiumByUID(uid string) bool
	GetChannelIDByType(channel string) (int64, error)
}

const (
	tokenLength   = 32
	maxTagLength  = 30
	minTextLength = 100
)

type token string

// generateToken генерирует уникальный криптоустойчивый токен
func (r ApiRepository) generateToken() token {
	// генерируем случайную строку заданной длины
	tokenBytes := make([]byte, tokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Printf("Ошибка генерации токена: %s", err)
		return ""
	}
	tokenValue := base64.URLEncoding.EncodeToString(tokenBytes)

	// хэшируем строку токена
	hashBytes := sha256.Sum256([]byte(tokenValue))
	hashValue := fmt.Sprintf("%x", hashBytes)

	resultToken := token(hashValue)

	return resultToken
}

func (r ApiRepository) GenerateUID(login, password string) string {
	// генерируем случайную строку заданной длины
	// объединяем login и password
	combinedString := login + password

	// хэшируем объединенную строку
	hashBytes := sha256.Sum256([]byte(combinedString))
	hashValue := fmt.Sprintf("%x", hashBytes)

	return hashValue
}

func (r ApiRepository) SaveTransaction(transaction model.Transaction) error {
	_, err := r.store.db.Query("INSERT INTO payment_transactions (uid, ext_id, client_id, type) VALUES ($1, $2, $3, $4)",
		transaction.UID,
		transaction.ExtID,
		transaction.ClientID,
		transaction.Type,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r ApiRepository) GetTransaction(uid string) (transaction model.Transaction, err error) {
	err = r.store.db.QueryRow("SELECT uid, ext_id, client_id, date, type, ended FROM payment_transactions WHERE ext_id=$1",
		uid,
	).Scan(
		&transaction.UID,
		&transaction.ExtID,
		&transaction.ClientID,
		&transaction.Date,
		&transaction.Type,
		&transaction.Ended,
	)
	if err != nil {
		return transaction, err
	}
	return transaction, nil
}

func (r ApiRepository) endTransaction(uid string) error {
	_, err := r.store.db.Query("UPDATE payment_transactions SET ended = true WHERE ext_id=$1",
		uid,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r ApiRepository) MakePremiumStatus(uid string, inviteLink string) error {
	transaction, err := r.GetTransaction(uid)
	if err != nil {
		return err
	}
	if transaction.Ended {
		return ErrTransaсtionAlreadyEnded
	}
	var months, payed int
	_, err = r.store.db.Query("INSERT INTO premium_subscribe (uid, date_of_last_payment, total_payed, link) VALUES ($1, $2, $3, $4)",
		transaction.ClientID,
		time.Now().AddDate(0, months, 0),
		payed,
		inviteLink,
	)
	if err != nil {
		if strings.Contains(err.Error(), "constraint failed") || strings.Contains(err.Error(), "ограничение уникальности") {
			_, err = r.store.db.Query("UPDATE premium_subscribe SET date_of_last_payment=$1, total_payed=total_payed+$2 WHERE uid=$3",
				time.Now(),
				1,
				transaction.ClientID,
			)
			log.Println("Зафиксирована повторная оплата uid: " + uid + " ext_id: " + transaction.ExtID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	err = r.endTransaction(uid)
	if err != nil {
		return err
	}

	return nil
}

func (r ApiRepository) GetPaymentRequestByUID(uid string) (model.Transaction, error) {
	var payment model.Transaction
	err := r.store.db.QueryRow("SELECT uid, ext_id, client_id, date, type, ended FROM payment_transactions WHERE uid=$1",
		uid,
	).Scan(
		&payment.UID,
		&payment.ExtID,
		&payment.ClientID,
		&payment.Date,
		&payment.Type,
		&payment.Ended,
	)
	if err != nil {
		return payment, err
	}
	return payment, nil
}

func (r ApiRepository) CheckPremiumByUID(uid string) bool {
	var expireDate time.Time
	err := r.store.db.QueryRow("SELECT expire_date FROM premium_subscribe WHERE uid=$1",
		uid,
	).Scan(
		&expireDate,
	)
	if err != nil {
		return false
	}
	if time.Now().After(expireDate) {
		return false
	}
	return true
}

func (r ApiRepository) GetChannelIDByType(channel string) (int64, error) {
	var id int64
	err := r.store.db.QueryRow("SELECT id FROM channels WHERE type=$1",
		channel,
	).Scan(
		&id,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}
