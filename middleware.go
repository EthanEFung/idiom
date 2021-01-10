package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func slackSecretMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)

		key := []byte(os.Getenv("SLACK_SECRET"))
		buf, err := ioutil.ReadAll(r.Body)
		if writeError(w, err) {
			return
		}
		reader := ioutil.NopCloser(bytes.NewBuffer(buf))
		v, t := "v0", r.Header.Get("X-Slack-Request-Timestamp")
		stamp, err := strconv.Atoi(t)
		if (time.Now().Unix() - int64(stamp)) > 60*5 {
			http.Error(w, "Time Expired", 403)
			return
		}
		message := []byte(fmt.Sprintf("%s:%s:%s", v, t, buf))
		mac := hmac.New(sha256.New, key)
		mac.Write(message)

		expected := []byte(v + "=" + hex.EncodeToString(mac.Sum(nil)))
		actual := []byte(r.Header.Get("X-Slack-Signature"))

		if hmac.Equal(expected, actual) {
			r.Body = reader
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", 403)
		}
	})
}
