package tokens

import (
	"errors"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoMaker struct{
	paseto *paseto.V2
	symmetricKey []byte
}

func NewToken(key string)(TokenMaker,error){
	if len(key) != chacha20poly1305.KeySize{
		return nil,errors.New("key must be 32 bytes")
	}

	maker := &PasetoMaker{
		paseto: paseto.NewV2(),
		symmetricKey: []byte(key),
	}
	return maker,nil
}

func(maker *PasetoMaker)CreateToken(username string,duration time.Duration)(string,error){
	payload:= NewPayload(username,duration)
	return maker.paseto.Encrypt(maker.symmetricKey,payload,nil)

}

func(maker *PasetoMaker) VerifyToken(token string)(*Payload,error){
	payload:=&Payload{}
	err:=maker.paseto.Decrypt(token,maker.symmetricKey,payload,nil)
	if err!=nil{
		return nil, err
	}
	err = VerifyPayload(payload)
	if err != nil {
		return nil, err
	}
	return payload,nil
}