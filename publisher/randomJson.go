package main

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"time"
	"yuriy_test/models"
)

func randomJson() ([]byte, error) {
	min, max := 0, 100
	gofakeit.Seed(time.Now().UnixNano())
	order := models.Order{
		OrderUid:    uuid.NewString(),
		TrackNumber: gofakeit.LoremIpsumWord(),
		Entry:       gofakeit.LoremIpsumWord(),
		Delivery: models.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: fmt.Sprintf("%s, %d", gofakeit.StreetName(), gofakeit.Number(min, max)),
			Region:  gofakeit.Country(),
			Email:   gofakeit.Email(),
		},
		Payment: models.Payment{
			Transaction:  uuid.NewString(),
			RequestId:    uuid.NewString(),
			Currency:     gofakeit.Currency().Short,
			Provider:     gofakeit.Company(),
			Amount:       gofakeit.Number(min, max),
			PaymentDt:    gofakeit.Number(min, max),
			Bank:         gofakeit.Company(),
			DeliveryCost: gofakeit.Number(min, max),
			GoodsTotal:   gofakeit.Number(min, max),
			CustomFee:    gofakeit.Number(min, max),
		},
		Items: []models.Item{{
			ChrtId:      gofakeit.Number(min, max),
			TrackNumber: uuid.NewString(),
			Price:       gofakeit.Number(min, max),
			Rid:         uuid.NewString(),
			Name:        gofakeit.Name(),
			Sale:        gofakeit.Number(min, max),
			Size:        gofakeit.NounCountable(),
			TotalPrice:  gofakeit.Number(min, max),
			NmId:        gofakeit.Number(min, max),
			Brand:       gofakeit.Company(),
			Status:      gofakeit.Number(min, max),
		}},
		Locale:            gofakeit.LanguageAbbreviation(),
		InternalSignature: gofakeit.LoremIpsumWord(),
		CustomerId:        uuid.NewString(),
		DeliveryService:   gofakeit.Company(),
		Shardkey:          gofakeit.LoremIpsumWord(),
		SmId:              gofakeit.Number(min, max),
		DateCreated:       gofakeit.Date(),
		OofShard:          gofakeit.LoremIpsumWord(),
	}
	utc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		fmt.Println(err.Error())
	}
	order.DateCreated = order.DateCreated.In(utc)
	byteJSon, err := json.Marshal(order)
	if err != nil {
		return []byte{}, err
	}
	return byteJSon, nil
}
