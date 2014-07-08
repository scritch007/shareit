package tools

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
)

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}


//Get the client ip address...
func GetClientAddr(r *http.Request) string{
	ip := r.Header.Get("X-Real-IP")
	if 0 == len(ip){
		ip = r.Header.Get("X-Forwarded-For")
		if 0 == len(ip){
			ip = r.RemoteAddr
		}
	}
	return ip
}
