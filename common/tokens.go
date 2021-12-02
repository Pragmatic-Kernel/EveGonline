package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/square/go-jose"
	"gorm.io/gorm"
)

const JWK = `{"alg":"RS256","e":"AQAB","kid":"JWT-Signature-Key","kty":"RSA","n":"nehPQ7FQ1YK-leKyIg-aACZaT-DbTL5V1XpXghtLX_bEC-fwxhdE_4yQKDF6cA-V4c-5kh8wMZbfYw5xxgM9DynhMkVrmQFyYB3QMZwydr922UWs3kLz-nO6vi0ldCn-ffM9odUPRHv9UbhM5bB4SZtCrpr9hWQgJ3FjzWO2KosGQ8acLxLtDQfU_lq0OGzoj_oWwUKaN_OVfu80zGTH7mxVeGMJqWXABKd52ByvYZn3wL_hG60DfDWGV_xfLlHMt_WoKZmrXT4V3BCBmbitJ6lda3oNdNeHUh486iqaL43bMR2K4TzrspGMRUYXcudUQ9TycBQBrUlT85NRY9TeOw","use":"sig"}`

func GetTokens(db *gorm.DB) (*[]Token, error) {
	tokens := &[]Token{}
	_ = db.Find(tokens)
	return tokens, nil
}

func RefreshToken(token *Token, clientId, secretKey string) error {
	fmt.Printf("Refreshing Token, expired at: %d.\n", token.Exp)
	fmt.Println(token.Exp)
	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", token.RefreshToken)
	req, _ := http.NewRequest("POST", EveApiTokenUrl, strings.NewReader(params.Encode()))
	req.Header.Add("Host", "login.eveonline.com")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientId, secretKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Status code %d in body\n", resp.StatusCode)
		return errors.New("invalid Status Code")
	}
	token_ := PreToken{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &token_)
	if err != nil {
		return err
	}
	fmt.Println(token_.AccessToken)
	fmt.Println(token_.RefreshToken)
	payload, err := GetTokenPayload(token_.AccessToken)
	if err != nil {
		return err
	}
	token.AccessToken = token_.AccessToken
	token.RefreshToken = token_.RefreshToken
	token.Exp = payload.Exp
	fmt.Printf("Done Refreshing Token, expiring at: %d.\n", token.Exp)
	return nil
}

func GetTokenPayload(tokenString string) (Payload, error) {
	object, err := jose.ParseSigned(tokenString)
	if err != nil {
		fmt.Println("Error parsing token:", err)
		return Payload{}, err
	}
	var d jose.JSONWebKey
	if err := json.Unmarshal([]byte(JWK), &d); err != nil {
		fmt.Println("Error umarshalling web key:", err)
		return Payload{}, err
	}
	output, err := object.Verify(d)
	if err != nil {
		fmt.Println("Error verifying token:", err)
		return Payload{}, err
	}
	var payload Payload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		fmt.Println("Error unmarshalling payload:", err)
		return Payload{}, err
	}
	return payload, nil
}
