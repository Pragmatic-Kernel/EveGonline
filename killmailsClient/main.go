package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"

	"github.com/Pragmatic-Kernel/EveGonline/common"
	"golang.org/x/term"
)

var endpoint string

const PKID = 260635334

func main() {
	endpoint = os.Getenv("EVE_KMSERVER_ENDPOINT")
	if endpoint == "" {
		panic("No endpoint, please set EVE_KMSERVER_ENDPOINT")
	}
	if len(os.Args) == 1 {
		kms, err := getKillmails()
		if err != nil {
			panic(err)
		}
		err = formatKillmails(kms)
		if err != nil {
			panic(err)
		}
	} else if len(os.Args) == 2 {
		kmID := os.Args[1]
		km, err := getKillmail(kmID)
		if err != nil {
			panic(err)
		}
		err = formatKillmail(km)
		if err != nil {
			panic(err)
		}
	}
}

func getKillmails() (*[]common.EnrichedKMShort, error) {
	req, err := http.NewRequest("GET", endpoint+"/killmails/", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing GET request: %w", err)
	}
	defer resp.Body.Close()
	kms := []common.EnrichedKMShort{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading GET request body: %w", err)
	}
	json.Unmarshal(body, &kms)
	return &kms, nil

}

func getKillmail(kmID string) (*common.EnrichedKM, error) {
	req, err := http.NewRequest("GET", endpoint+"/killmail/"+kmID, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing GET request: %w", err)
	}
	defer resp.Body.Close()
	km := common.EnrichedKM{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading GET request body: %w", err)
	}
	json.Unmarshal(body, &km)
	return &km, nil
}

func formatKillmails(kms *[]common.EnrichedKMShort) error {
	fmt.Printf("%4s %15s %4s %25s %50s %25s %10s %9s\n", "KM", "System", "Sec.", "Victim", "Ship", "Final Blow", "Date", "ID")
	for i := 0; i < 149; i++ {
		fmt.Printf("=")
	}
	fmt.Printf("\n")
	for index, km := range *kms {
		kmDate := km.KillmailTime.Format("02/01/2006")
		loss := getKillmailStatus(&km)
		if loss {
			fmt.Printf("\033[31m#%-3d %15s %4.1f %25s %50s %25s %10s %9d\n", index, km.SolarSystem.Name, km.SolarSystem.SecurityStatus, km.Victim.CharacterName, km.Victim.ShipTypeName, km.Attacker.CharacterName, kmDate, km.ID)
		} else {
			fmt.Printf("\033[32m#%-3d %15s %4.1f %25s %50s %25s %10s %9d\n", index, km.SolarSystem.Name, km.SolarSystem.SecurityStatus, km.Victim.CharacterName, km.Victim.ShipTypeName, km.Attacker.CharacterName, kmDate, km.ID)
		}
	}
	return nil
}

func formatKillmail(km *common.EnrichedKM) error {
	width, _, err := term.GetSize(0)
	if err != nil {
		return fmt.Errorf("error getting term width: %w", err)
	}
	for i := 0; i < width; i++ {
		fmt.Printf("=")
	}
	fmt.Printf("\n")
	finalBlow := filterAttackers(km.Attackers)
	fmt.Printf("%s flying a %s in %s got killed by %s flying a %s\n", km.Victim.CharacterName, km.Victim.ShipTypeName, km.SolarSystem.Name, finalBlow.CharacterName, finalBlow.ShipTypeName)
	for i := 0; i < width; i++ {
		fmt.Printf("=")
	}
	fmt.Printf("\n")
	fmt.Printf("ATTACKERS:\n")
	for i := 0; i < 10; i++ {
		fmt.Printf("-")
	}
	fmt.Printf("\n")
	for _, attacker := range *km.Attackers {
		fmt.Printf("%-30s %-50s %-50s\n", attacker.CharacterName, attacker.ShipTypeName, attacker.WeaponTypeName)
	}
	fmt.Printf("ITEMS:\n")
	for i := 0; i < 10; i++ {
		fmt.Printf("-")
	}
	fmt.Printf("\n")
	enrichedItems := filterItems(km.Victim.EnrichedItems)
	for name, value := range enrichedItems {
		fmt.Printf("%-50s \033[32m%-10d \033[31m%-10d\033[39m\n", name, value["dropped"], value["destroyed"])
	}
	return nil
}

func filterAttackers(attackers *[]common.EnrichedAttacker) *common.EnrichedAttacker {
	for _, attacker := range *attackers {
		if attacker.FinalBlow {
			return &attacker
		}
	}
	attackers_ := *attackers
	return &attackers_[0]
}

func filterItems(items *[]common.EnrichedItem) map[string]map[string]uint {
	enrichedItems := make(map[string]map[string]uint)
	for _, item := range *items {
		if item_, ok := enrichedItems[item.ItemName]; ok {
			if item.QuantityDropped != 0 {
				item_["dropped"] += item.QuantityDropped
			} else {
				item_["destroyed"] += item.QuantityDestroyed
			}
		} else {
			if item.QuantityDropped != 0 {
				enrichedItems[item.ItemName] = make(map[string]uint)
				enrichedItems[item.ItemName]["dropped"] = item.QuantityDropped
			} else {
				enrichedItems[item.ItemName] = make(map[string]uint)
				enrichedItems[item.ItemName]["destroyed"] = item.QuantityDestroyed
			}
		}
	}
	return enrichedItems
}

func getKillmailStatus(km *common.EnrichedKMShort) bool {
	if km.Victim.CorporationID == PKID {
		return true
	}
	return false
}

func formatSolarSystemSecurity(security float64) string {
	fmt.Println(security)
	roundedStatus := math.Round(security/10.0) * 10.0
	fmt.Println(roundedStatus)
	return fmt.Sprintf("(%f)", roundedStatus)
}
