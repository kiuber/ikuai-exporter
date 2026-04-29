package ikuai

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jakeslee/ikuai-exporter/ikuai/action"
)

type IKuai struct {
	client   *resty.Client
	debug    bool
	Url      string
	Username string
	Password string

	session string
	IsV4    bool
}

func NewIKuai(url string, username string, password string, insecureSkipVerify, autoLogin bool) *IKuai {
	client := resty.New()

	if insecureSkipVerify {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	i := &IKuai{
		client:   client,
		Url:      url,
		Username: username,
		Password: password,
	}

	if autoLogin {
		client.SetRetryCount(3)
		client.SetRetryWaitTime(5 * time.Second)
		client.AddRetryCondition(func(response *resty.Response, err error) bool {
			body := response.Body()
			var result action.Result
			rErr := json.Unmarshal(body, &result)
			if rErr != nil {
				log.Printf("Unmarshal error: %v, username: %s, body: %s", rErr, username, body)
				return false
			}

			isTimeout := result.Result == 10014

			if !isTimeout {
				var v4 struct {
					Code int `json:"code"`
				}
				if json.Unmarshal(body, &v4) == nil && v4.Code == 1008 {
					isTimeout = true
				}
			}

			if isTimeout {
				log.Printf("session timeout: try to login")
				_, err := i.Login()
				if err != nil {
					return false
				}

				log.Printf("successfully login, re-try to meter states")

				return true
			}

			return false
		})
	}

	return i
}

type LoginRequest struct {
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
}

func getMD5(password string) string {
	hash := md5.New()
	hash.Write([]byte(password))
	sum := hash.Sum(nil)

	return fmt.Sprintf("%x", sum)
}

func (i *IKuai) Login() (string, error) {
	var result action.Result

	response, err := i.client.R().
		SetBody(&LoginRequest{
			Username: i.Username,
			Passwd:   getMD5(i.Password),
		}).
		SetResult(&result).
		Post(i.Url + "/Action/login")

	if err != nil {
		return "", err
	}

	for _, cookie := range response.Cookies() {
		if cookie.Name == "sess_key" {
			i.session = cookie.Value
			return cookie.Value, nil
		}
	}

	log.Printf("login error: %s", response.Body())

	return "", errors.New(fmt.Sprintf("login error: %s, no cookies", result.ErrMsg))
}

func normalizeResponse(body []byte) []byte {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return body
	}

	changed := false

	if _, ok := raw["results"]; ok {
		if _, ok := raw["Data"]; !ok {
			raw["Data"] = raw["results"]
			delete(raw, "results")
			changed = true
		}
	}

	if _, ok := raw["code"]; ok {
		if _, ok := raw["Result"]; !ok {
			raw["Result"] = raw["code"]
			delete(raw, "code")
			changed = true
		}
	}

	if _, ok := raw["message"]; ok {
		if _, ok := raw["ErrMsg"]; !ok {
			raw["ErrMsg"] = raw["message"]
			delete(raw, "message")
			changed = true
		}
	}

	if !changed {
		return body
	}

	normalized, err := json.Marshal(raw)
	if err != nil {
		return body
	}

	return normalized
}

func (i *IKuai) Run(session string, act *action.Action, result interface{}) (string, error) {
	url := i.Url + "/Action/call"

	response, err := i.client.R().
		SetHeader("Content-Type", "application/json").
		SetCookie(&http.Cookie{Name: "sess_key", Value: session}).
		SetBody(act).
		Post(url)

	if err != nil {
		return "", err
	}

	body := normalizeResponse(response.Body())

	if err := json.Unmarshal(body, result); err != nil {
		return "", err
	}

	if i.debug {
		log.Printf("POST %s, request: %v, response: %s", url, act, string(body))
	}

	return string(body), nil
}

func (i *IKuai) DetectVersion() error {
	resp, err := i.ShowSysStat()
	if err != nil {
		return err
	}

	if resp != nil && action.IsV4(resp.Data.SysStat.Verinfo.Version) {
		i.IsV4 = true
		log.Printf("detected ikuai v4: %s", resp.Data.SysStat.Verinfo.Version)
	}

	return nil
}

func (i *IKuai) Debug() {
	i.debug = true
}
