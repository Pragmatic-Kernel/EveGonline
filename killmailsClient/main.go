package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"

	"github.com/Pragmatic-Kernel/EveGonline/common"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

var endpoint string

const PKID = 260635334

func main() {
	endpoint = os.Getenv("EVE_KMSERVER_ENDPOINT")
	if endpoint == "" {
		fmt.Println("No endpoint, please set EVE_KMSERVER_ENDPOINT")
		return
	}
	if len(os.Args) == 1 {
		kms, err := getKillmails()
		if err != nil {
			panic(err)
		}
		items := []list.Item{}
		for _, km := range *kms {
			items = append(items, item(km))
		}

		width, height, err := term.GetSize(0)
		if err != nil {
			panic(err)
		}

		l := list.New(items, itemDelegate{}, width, height-5)
		l.Title = "Pragmatic Kernel Killmails"
		l.SetShowStatusBar(true)
		l.SetFilteringEnabled(true)
		l.Styles.Title = titleStyle
		l.Styles.PaginationStyle = paginationStyle
		l.Styles.HelpStyle = helpStyle

		m := model{list: l}

		if err := tea.NewProgram(m).Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	} else if len(os.Args) == 2 {
		kmID := os.Args[1]
		km, err := getKillmail(kmID)
		if err != nil {
			fmt.Println("No killmail found with id: ", kmID)
			return
		}
		kmString, err := formatKillmail(km)
		if err != nil {
			fmt.Println("Error while formatting killmail: ", err)
			return
		}
		fmt.Println(kmString)
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request status error: %d", resp.StatusCode)
	}
	km := common.EnrichedKM{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading GET request body: %w", err)
	}
	json.Unmarshal(body, &km)
	return &km, nil
}

func formatKillmails(kms *[]common.EnrichedKMShort) ([]string, error) {
	result := []string{}
	for _, km := range *kms {
		res := ""
		kmDate := km.KillmailTime.Format("02/01/2006")
		loss := getKillmailStatus(&km)
		if loss {
			res = fmt.Sprintf("\033[31m %15s %4.1f %25s %50s %25s %10s %9d \033[0m", km.SolarSystem.Name, km.SolarSystem.SecurityStatus, km.Victim.CharacterName, km.Victim.ShipTypeName, km.Attacker.CharacterName, kmDate, km.ID)
		} else {
			res = fmt.Sprintf("\033[32m %15s %4.1f %25s %50s %25s %10s %9d \033[0m", km.SolarSystem.Name, km.SolarSystem.SecurityStatus, km.Victim.CharacterName, km.Victim.ShipTypeName, km.Attacker.CharacterName, kmDate, km.ID)
		}
		result = append(result, res)
	}
	return result, nil
}

func formatKillmailShort(km *common.EnrichedKMShort) string {
	res := ""
	kmDate := km.KillmailTime.Format("02/01/2006")
	loss := getKillmailStatus(km)
	if loss {
		res = fmt.Sprintf("\033[31m %15s %4.1f %25s %50s %25s %10s\033[0m", km.SolarSystem.Name, km.SolarSystem.SecurityStatus, km.Victim.CharacterName, km.Victim.ShipTypeName, km.Attacker.CharacterName, kmDate)
	} else {
		res = fmt.Sprintf("\033[32m %15s %4.1f %25s %50s %25s %10s\033[0m", km.SolarSystem.Name, km.SolarSystem.SecurityStatus, km.Victim.CharacterName, km.Victim.ShipTypeName, km.Attacker.CharacterName, kmDate)
	}
	return res
}

func formatKillmail(km *common.EnrichedKM) (string, error) {
	width, _, err := term.GetSize(0)
	res := ""
	if err != nil {
		return "", fmt.Errorf("error getting term width: %w", err)
	}
	for i := 0; i < width; i++ {
		res += "="
	}
	res += "\n"
	finalBlow := filterAttackers(km.Attackers)
	res += fmt.Sprintf("%s flying a %s in %s got killed by %s flying a %s\n", km.Victim.CharacterName, km.Victim.ShipTypeName, km.SolarSystem.Name, finalBlow.CharacterName, finalBlow.ShipTypeName)
	for i := 0; i < width; i++ {
		res += "="
	}
	res += "\n"
	res += "ATTACKERS:\n"
	for i := 0; i < 10; i++ {
		res += "-"
	}
	res += "\n"
	for _, attacker := range *km.Attackers {
		res += fmt.Sprintf("%-30s %-50s %-50s\n", attacker.CharacterName, attacker.ShipTypeName, attacker.WeaponTypeName)
	}
	res += "\n"
	res += "ITEMS:\n"
	for i := 0; i < 6; i++ {
		res += "-"
	}
	res += "\n"
	enrichedItems := filterItems(km.Victim.EnrichedItems)
	for name, value := range enrichedItems {
		res += fmt.Sprintf("%-50s \033[32m%-10d \033[31m%-10d\033[39m\n", name, value["dropped"], value["destroyed"])
	}
	return res, nil
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
	return km.Victim.CorporationID == PKID
}

func formatSolarSystemSecurity(security float64) string {
	roundedStatus := math.Round(security/10.0) * 10.0
	return fmt.Sprintf("(%f)", roundedStatus)
}
