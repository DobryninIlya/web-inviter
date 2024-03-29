package payments

import "time"

type YokassaPayment struct {
	Amount       Amount       `json:"amount"`
	Capture      bool         `json:"capture"`
	Confirmation Confirmation `json:"confirmation"`
	Description  string       `json:"description"`
	Receipt      Receipt      `json:"receipt"`
}

type Receipt struct {
	Items    []Items  `json:"items"`
	Customer Customer `json:"customer"`
}

type Items struct {
	Description string `json:"description"`
	Amount      Amount `json:"amount"`
	VatCode     int    `json:"vat_code"`
	Quantity    string `json:"quantity"`
}

type Customer struct {
	Email string `json:"email"`
}

type Confirmation struct {
	Type      string `json:"type"`
	ReturnUrl string `json:"return_url"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Recipient struct {
	AccountId string `json:"account_id"`
	GatewayId string `json:"gateway_id"`
}

type Payment struct {
	Id           string       `json:"id"`
	Status       string       `json:"status"`
	Amount       Amount       `json:"amount"`
	Description  string       `json:"description"`
	Recipient    Recipient    `json:"recipient"`
	CreatedAt    time.Time    `json:"created_at"`
	Confirmation Confirmation `json:"confirmation"`
	Test         bool         `json:"test"`
	Paid         bool         `json:"paid"`
	Refundable   bool         `json:"refundable"`
	Metadata     struct {
	} `json:"metadata"`
}

type Notification struct {
	Type   string  `json:"type"`
	Event  string  `json:"event"`
	Object Payment `json:"object"`
}
