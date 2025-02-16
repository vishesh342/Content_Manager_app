package tokens

import (
	"errors"
	"time"
)

type TokenMaker interface{
	CreateToken(string, time.Duration)(string,error)
	VerifyToken(string)(*Payload,error)
}

type Payload struct{
	Username string `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(username string,duration time.Duration)(*Payload){
	return &Payload{
		Username: username,
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
}

func VerifyPayload(req *Payload)(error){
	if time.Now().After(req.ExpiredAt){
		return errors.New("token expired")
	}
	return nil
}