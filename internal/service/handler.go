package service

import (
	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/api/gen/v1"
)

type SecraHandler struct {
	secra_v1.UnimplementedSecraServiceServer
	DB *bun.DB
}

func NewSecraHandler(db *bun.DB) *SecraHandler {
	return &SecraHandler{DB: db}
}
