package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Pragmatic-Kernel/EveGoNline/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ClientId string
var SecretKey string
var CallbackUri string

func getCharID(charID string) (uint, error) {
	segments := strings.Split(charID, ":")
	if len(segments) < 3 {
		errMsg := fmt.Sprintf("unable to extract CharID from %s", charID)
		return 0, errors.New(errMsg)
	}
	charIDint, err := strconv.Atoi(segments[2])
	if err != nil {
		return 0, err
	}
	return uint(charIDint), nil

}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&common.Token{})
	ClientId = os.Getenv("CLIENT_ID")
	SecretKey = os.Getenv("SECRET_KEY")
	CallbackUri = os.Getenv("CALLBACK_URI")
	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, _ *http.Request) {
		params := url.Values{}
		params.Add("response_type", "code")
		params.Add("redirect_uri", CallbackUri)
		params.Add("client_id", ClientId)
		params.Add("state", randomString(16))
		location := common.EveApiAuthorizeUrl + "?" + params.Encode()
		//FIXME hardcoded scopes
		location += "&scope=esi-killmails.read_killmails.v1%20esi-killmails.read_corporation_killmails.v1"
		fmt.Println(location)
		w.Header().Add("Location", location)
		w.WriteHeader(301)
		w.Write([]byte{})
	})
	mux.HandleFunc("/callback/", func(_ http.ResponseWriter, r *http.Request) {
		codes, ok := r.URL.Query()["code"]
		if !ok {
			fmt.Println("ERROR: no code found in URL")
			return
		}
		code := codes[0]
		params := url.Values{}
		params.Add("grant_type", "authorization_code")
		params.Add("code", code)
		req, err := http.NewRequest("POST", common.EveApiTokenUrl, strings.NewReader(params.Encode()))
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}
		req.Header.Add("Host", "login.eveonline.com")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(ClientId, SecretKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}
		pretoken := common.PreToken{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}
		err = json.Unmarshal(body, &pretoken)
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}
		fmt.Println(pretoken.AccessToken)
		fmt.Println(pretoken.RefreshToken)
		payload, err := common.GetTokenPayload(pretoken.AccessToken)
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}
		charID, err := getCharID(payload.Sub)
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}
		token := common.Token{}
		token.AccessToken = pretoken.AccessToken
		token.RefreshToken = pretoken.RefreshToken
		token.Exp = payload.Exp
		result := db.Create(&token)
		if result.Error != nil {
			fmt.Println("ERROR:", err)
			return
		}
		fmt.Printf("Added Token for charID %d", charID)
	})
	s := &http.Server{
		Addr:    ":4200",
		Handler: mux,
	}
	s.ListenAndServe()
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
