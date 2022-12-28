package tools

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
)

const SessionPrefix = "sess_"

func GetRandString(length int) string {
	r := make([]byte, length)
	io.ReadFull(rand.Reader, r)
	return base64.URLEncoding.EncodeToString(r)
}

func CreateSessionId(sessionId string) string {
	return SessionPrefix + sessionId
}

func Md5(content string) (hashString string) {
	h := md5.New()
	io.WriteString(h, content)
	return hex.EncodeToString(h.Sum(nil))
}

func GetSessionName(sessionID string) string {
	return SessionPrefix + sessionID
}
