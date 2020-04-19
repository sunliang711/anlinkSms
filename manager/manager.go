package manager

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

type AnlinkSmsManger struct {
	APIKey      string
	APISecret   string
	TaskCode    string
	ChannelType string
	SmsURL      string
	ContentType string
	Accept      string
}

func NewAnlinkSmsManager(u string, key string, secret string, taskCode string, channelType string) *AnlinkSmsManger {
	return &AnlinkSmsManger{
		APIKey:      key,
		APISecret:   secret,
		TaskCode:    taskCode,
		ChannelType: channelType,
		SmsURL:      u,
		ContentType: "application/json; charset=utf-8",
		Accept:      "application/json",
	}
}

type RequestBody struct {
	TaskCode    string      `json:"taskCode"`
	ChannelType string      `json:"channelType"`
	Receiver    string      `json:"receiver"`
	Params      interface{} `json:"params"`
}

func (man *AnlinkSmsManger) Send(receiver string, data map[string]string) error {
	logrus.SetLevel(logrus.DebugLevel)

	body := RequestBody{
		TaskCode:    man.TaskCode,
		ChannelType: man.ChannelType,
		Receiver:    receiver,
		Params:      data,
	}
	bs, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", man.SmsURL, bytes.NewReader(bs))
	if err != nil {
		return err
	}

	// build headers
	var headers sortedArray

	headers.Add("x-anlink-apikey", man.APIKey)
	now := time.Now().Unix()
	headers.Add("datetime", fmt.Sprintf("%d", now))
	rand.Seed(now)
	headers.Add("x-anlink-signature-nonce", fmt.Sprintf("%d", rand.Int31()))
	headers.Add("prd-id", "0")
	headers.Add("x-anlink-if-verify-response", "1")
	headers.Add("x-anlink-if-sign-postbody", "1")
	headers.Add("content-type", man.ContentType)
	headers.Add("accept", man.Accept)
	// TODO check if sign postbody

	sort.Sort(headers)
	data2Signed, err := headers.ToJsonObject()
	if err != nil {
		return err
	}
	logrus.Debugf("headers: %v\n", string(data2Signed))
	signature, err := sign(data2Signed, []byte(man.APISecret))
	if err != nil {
		return err
	}
	logrus.Debugf("Signature: %x\n", signature)
	headers.Add("x-anlink-signature", string(signature))

	for _, kv := range headers {
		k := kv[0].(string)
		v := kv[1].(string)
		request.Header.Add(k, v)
	}

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// HMAC SHA1
func sign(data []byte, key []byte) ([]byte, error) {
	mac := hmac.New(sha1.New, key)
	mac.Write(data)
	return mac.Sum(nil), nil
}
