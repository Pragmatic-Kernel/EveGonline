package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Pragmatic-Kernel/EveGoNline/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ClientId string
var SecretKey string

func main() {
	ClientId = os.Getenv("CLIENT_ID")
	SecretKey = os.Getenv("SECRET_KEY")
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	db.AutoMigrate(&common.Mapping{}, &common.Token{}, &common.Killmail{}, &common.Attacker{}, &common.Victim{}, &common.Item{}, &common.SubItem{}, &common.Position{}, &common.SolarSystem{})
	db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		panic(err)
	}
	for {
		tokens, err := common.GetTokens(db)
		fmt.Printf("Found %d tokens\n", len(*tokens))
		if err != nil {
			panic(err)
		}
		for _, token := range *tokens {
			mappings, err := common.GetMappings(db)
			fmt.Printf("Found %d mappings\n", len(mappings))
			if err != nil {
				panic(err)
			}
			unknownIDs := []uint{}
			existingKms := []common.Killmail{}
			db.Select("id").Find(&existingKms)
			existingKmIds := getExistingKmIds(&existingKms)
			fmt.Printf("Found %d Killmails\n", len(existingKmIds))
			fmt.Printf("Retrieving KM IDs for token: %d\n", token.ID)
			newKms, err := getKillmailIDsWithToken(db, token)
			if err != nil {
				fmt.Println("Error while retrieving KM IDs:", err)
				continue
			}
			filteredKms := []common.Killmail{}
			for _, km := range newKms {
				if _, ok := existingKmIds[km.ID]; ok {
					continue
				} else {
					filteredKms = append(filteredKms, km)
				}
			}
			fmt.Printf("Killmails post filtering: %d\n", len(filteredKms))
			if len(filteredKms) > 10 {
				filteredKms = filteredKms[:10]
			}
			KMsToCreate := []common.Killmail{}
			for _, km := range filteredKms {
				err := getKillmailDetails(&km)
				if err != nil {
					fmt.Println("Error while retrieving KM Details:", err)
					continue
				}
				KMsToCreate = append(KMsToCreate, km)
				unknownIDsKM := getUnknownIDs(&km, mappings)
				unknownIDs = append(unknownIDs, unknownIDsKM...)
				if len(unknownIDs) > 100 {
					fmt.Printf("We have %d unkown items to retrieve, breaking loop.\n", len(unknownIDs))
					break
				}
			}
			if len(unknownIDs) > 0 {
				IDsmappings, err := retrieveUnknownIDs(unknownIDs)
				if err != nil {
					fmt.Printf("Error while retrieving unknownIDs, skipping.\n")
				} else {
					db.Create(IDsmappings)
				}
			}
			if len(KMsToCreate) > 0 {
				db.Create(&KMsToCreate)
			} else {
				fmt.Println("No killmails to save, skipping.")
			}
			fmt.Printf("Token %d done. Sleeping for %d minute.\n", token.ID, 1)
			time.Sleep(1 * time.Minute)
		}
		fmt.Printf("All tokens done. Sleeping for %d minutes.\n", 60)
		time.Sleep(60 * time.Minute)
	}
}

func getKillmailIDsWithToken(db *gorm.DB, token common.Token) ([]common.Killmail, error) {
	res := []common.Killmail{}
	var url string
	if token.CorpID != 0 {
		url = fmt.Sprintf(common.EveApiKillmailCorpAPIUrl, token.CorpID)
		fmt.Println(url)
	} else {
		url := fmt.Sprintf(common.EveApiKillmailCharAPIUrl, token.CharID)
		fmt.Println(url)
	}
	body, err := getCache(url, 86400)
	if err != nil {
		return nil, fmt.Errorf("error fetching cache: %w", err)
	}
	if body != nil {
		if err := json.Unmarshal(body, &res); err != nil {
			return []common.Killmail{}, err
		}
		return res, nil
	}
	if int64(token.Exp) < time.Now().Unix() {
		err := common.RefreshToken(&token, ClientId, SecretKey)
		if err != nil {
			fmt.Println("Error while refreshing token:", err)
			return nil, err
		}
		db.Save(&token)
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	req.Header.Add("User-Agent", "CharName: Laszlo Bariani")
	if err != nil {
		return res, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return res, fmt.Errorf("invalid Status Code: %s", resp.Status)
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}
	body, err = setCache(url, body)
	if err != nil {
		return res, err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return []common.Killmail{}, err
	}
	return res, nil
}

func getExistingKmIds(kms *[]common.Killmail) map[uint]bool {
	res := make(map[uint]bool)
	for _, km := range *kms {
		res[km.ID] = true
	}
	return res
}

func getKillmailDetails(km *common.Killmail) error {
	id := km.ID
	hash := km.Hash
	url := fmt.Sprintf(common.EveApiKillmailDetailsAPIUrl, id, hash)
	body, err := getCache(url, 0)
	if err != nil {
		return fmt.Errorf("error fetching cache: %w", err)
	}
	if body != nil {
		if err := json.Unmarshal(body, &km); err != nil {
			return err
		}
		return nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", "CharName: Laszlo Bariani")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Status Code: %d\n", resp.StatusCode)
		return fmt.Errorf("invalid Status Code: %s", resp.Status)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body, err = setCache(url, body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, km); err != nil {
		return err
	}
	fmt.Printf("KM %d done. Sleeping for %d seconds.\n", km.ID, 10)
	time.Sleep(10 * time.Second)
	return nil
}

func getUnknownIDs(km *common.Killmail, mapping map[uint]string) []uint {
	res := []uint{}
	for _, attacker := range *km.Attackers {
		if _, ok := mapping[attacker.CharacterID]; !ok {
			res = append(res, attacker.CharacterID)
		}
		if _, ok := mapping[attacker.CorporationID]; !ok {
			res = append(res, attacker.CorporationID)
		}
		if _, ok := mapping[attacker.ShipTypeID]; !ok {
			res = append(res, attacker.ShipTypeID)
		}
		if _, ok := mapping[attacker.WeaponTypeID]; !ok {
			res = append(res, attacker.WeaponTypeID)
		}
	}
	if _, ok := mapping[km.Victim.CharacterID]; !ok {
		res = append(res, km.Victim.CharacterID)
	}
	if _, ok := mapping[km.Victim.CorporationID]; !ok {
		res = append(res, km.Victim.CorporationID)
	}
	if km.Victim.Items != nil {
		for _, item := range *km.Victim.Items {
			if _, ok := mapping[item.ItemTypeID]; !ok {
				res = append(res, item.ItemTypeID)
			}
			if item.SubItems != nil {
				for _, subitem := range *item.SubItems {
					if _, ok := mapping[subitem.ItemTypeID]; !ok {
						res = append(res, subitem.ItemTypeID)
					}
				}
			}
		}
	}
	res = filterUnknownIDs(res)
	return res
}

func retrieveUnknownIDs(unknownIDs []uint) (*[]common.Mapping, error) {
	mappings := []common.Mapping{}
	fmt.Printf("Need to retrieve %d IDs.\n", len(unknownIDs))
	unknownIDs = filterUnknownIDs(unknownIDs)
	fmt.Printf("Need to retrieve %d filtered IDs.\n", len(unknownIDs))
	if len(unknownIDs) > 200 {
		return nil, errors.New("too many IDs, skipping")
	}
	IDsList, err := json.Marshal(unknownIDs)
	if err != nil {
		return nil, err
	}
	fmt.Println("To retrieve:", string(IDsList))
	req, err := http.NewRequest("POST", common.EveApiNamesAPIUrl, bytes.NewReader(IDsList))
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "CharName: Laszlo Bariani")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Status code %d in body\n", resp.StatusCode)
		return nil, errors.New("invalid Status Code")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &mappings)
	if err != nil {
		return nil, err
	}
	return &mappings, nil
}

func filterUnknownIDs(unknownIDs []uint) []uint {
	res := []uint{}
	for _, elem := range unknownIDs {
		if elem == 0 || contains(elem, res) {
			continue
		}
		res = append(res, elem)
	}
	return res
}

func contains(id uint, ids []uint) bool {
	for _, elem2 := range ids {
		if id == elem2 {
			return true
		}
	}
	return false
}
