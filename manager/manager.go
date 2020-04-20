package manager

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	SignPostBodyY = iota
	SignPostBodyN
)

type smsResp struct {
	Code    string `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

type AnlinkSmsManger struct {
	APIKey        string
	APISecret     string
	TaskCode      string
	ChannelType   string
	SmsURL        string
	ContentType   string
	Accept        string
	SignPostBody  int
	AnlinkVersion string
}

func NewAnlinkSmsManager(u string, key string, secret string, taskCode string, channelType string) *AnlinkSmsManger {
	return &AnlinkSmsManger{
		APIKey:        key,
		APISecret:     secret,
		TaskCode:      taskCode,
		ChannelType:   channelType,
		SmsURL:        u,
		ContentType:   "application/json; charset=utf-8",
		Accept:        "application/json",
		SignPostBody:  SignPostBodyN,
		AnlinkVersion: "2",
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

	logrus.Debugf("request body: %v", string(bs))
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
	headers.Add("x-anlink-if-sign-postbody", fmt.Sprintf("%d", man.SignPostBody))
	headers.Add("content-type", man.ContentType)
	headers.Add("accept", man.Accept)
	headers.Add("x-anlink-signature-method", "HMAC-SHA1")
	headers.Add("x-anlink-version", man.AnlinkVersion)
	if man.SignPostBody == SignPostBodyY {
		// TODO sign postbody

	}

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
	logrus.Debugf("Signature: %v\n", signature)
	headers.Add("x-anlink-signature", signature)

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

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logrus.Debugf("sms resp: %v", string(respBytes))

	var sResp smsResp
	err = json.Unmarshal(respBytes, &sResp)
	if err != nil {
		return fmt.Errorf("Decode sms resp error: %v", err)
	}
	if sResp.Success {
		return nil
	}
	return fmt.Errorf("error: %v", sResp.Message)

}

// HMAC SHA1
func sign(data []byte, key []byte) (string, error) {
	mac := hmac.New(sha1.New, key)
	mac.Write(data)
	// return mac.Sum(nil), nil
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
