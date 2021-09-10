package server

import "encoding/json"

type BalanceRequest struct {
	AccountId int `json:"id"`
}

func (e BalanceRequest) Marshal() []byte {
	return noErrsMarshal(&e)
}

type BalanceResponse struct {
	CentsSum int64 `json:"cents_sum"`
}

func (e BalanceResponse) Marshal() []byte {
	return noErrsMarshal(&e)
}

type AddRequest struct {
	AccountId int   `json:"id"`
	CentsSum  int64 `json:"cents_sum"`
}

func (e AddRequest) Marshal() []byte {
	return noErrsMarshal(&e)
}

type WithDrawRequest struct {
	AccountId int   `json:"id"`
	CentsSum  int64 `json:"cents_sum"`
}

func (e WithDrawRequest) Marshal() []byte {
	return noErrsMarshal(&e)
}

type TransferRequest struct {
	SenderId    int   `json:"sender_id"`
	RecipientId int   `json:"recipient_id"`
	CentsSum    int64 `json:"cents_sum"`
}

func (e TransferRequest) Marshal() []byte {
	return noErrsMarshal(&e)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (e ErrorResponse) Marshal() []byte {
	return noErrsMarshal(&e)
}

func noErrsMarshal(v interface{}) []byte {
	bs, _ := json.Marshal(v)
	return bs
}
