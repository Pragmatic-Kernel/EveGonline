package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/Pragmatic-Kernel/EveGonline/common"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

var endpoint string
var debug *bool

const PKID = 260635334

func main() {
	debug = flag.Bool("d", false, "debug")
	flag.Parse()
	endpoint = os.Getenv("EVE_KMSERVER_ENDPOINT")
	if endpoint == "" {
		fmt.Println("No endpoint, please set EVE_KMSERVER_ENDPOINT")
		return
	}
	if *debug {
		f, _ := os.OpenFile("file.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer f.Close()
		log.SetOutput(f)
	}

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

	if err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
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

func formatKillmailShort(km *common.EnrichedKMShort) string {
	res := ""
	kmDate := km.KillmailTime.Format("02/01/2006 15:04:05")
	if km.Victim.CharacterID == 0 {
		km.Victim.CharacterName = km.Victim.CorporationName
	}
	if km.Attacker.CharacterID == 0 {
		km.Attacker.CharacterName = km.Attacker.ShipTypeName
	}
	loss := getKillmailStatus(km)
	if loss {
		res = fmt.Sprintf("\033[31m \033[3m%15s %4.1f\033[23m %25s \033[1m%50s\033[22m %25s \033[1m%15s\033[22m \033[3m%25s\033[22m\033[0m", km.SolarSystem.Name, km.SolarSystem.SecurityStatus, km.Victim.CharacterName, km.Victim.ShipTypeName, km.Attacker.CharacterName, common.FormatPrice(km.Price), kmDate)
	} else {
		res = fmt.Sprintf("\033[32m \033[3m%15s %4.1f\033[23m %25s \033[1m%50s\033[22m %25s \033[1m%15s\033[22m \033[3m%25s\033[22m\033[0m", km.SolarSystem.Name, km.SolarSystem.SecurityStatus, km.Victim.CharacterName, km.Victim.ShipTypeName, km.Attacker.CharacterName, common.FormatPrice(km.Price), kmDate)
	}
	return res
}

func formatKillmail(km *common.EnrichedKM) (string, error) {
	res := ""
	finalBlow := filterAttackers(km.Attackers)
	if finalBlow.CharacterID == 0 {
		finalBlow.CharacterName = finalBlow.ShipTypeName
	}
	res += fmt.Sprintf("\033[1m\033[31m%s\033[39m\033[22m lost a \033[1m%s\033[22m in \033[3m%s (%.1f)\033[23m. Final blow: \033[1m\033[32m%s\033[39m\033[22m\n", km.Victim.CharacterName, km.Victim.ShipTypeName, km.SolarSystem.Name, km.SolarSystem.SecurityStatus, finalBlow.CharacterName)
	res += "\n"
	res += fmt.Sprintf("\033[1m\033[33mKill Value: %s\033[39m\033[22m\n", common.FormatPrice(km.Price))
	res += "\n"
	res += "\n"
	res += fmt.Sprintf("\033[1m\033[4mAttackers: %d\033[22m\033[24m\n", len(*km.Attackers))
	res += "\n"
	for _, attacker := range *km.Attackers {
		if attacker.CharacterID == 0 {
			attacker.CharacterName = attacker.ShipTypeName
		}
		if attacker.WeaponTypeID == 0 {
			attacker.WeaponTypeName = attacker.ShipTypeName
		}
		res += fmt.Sprintf("\033[1m%50s\033[22m %50s \033[3m%50s %10d %5.1f%%\033[23m\n", attacker.CharacterName, attacker.ShipTypeName, attacker.WeaponTypeName, attacker.DamageDone, getDamagePercent(attacker.DamageDone, km.Victim.DamageTaken))
	}
	res += "\n"
	res += "\n"
	res += "\033[1m\033[4mItems:\033[22m\033[24m\n"
	res += "\n"
	res += fmt.Sprintf("\033[1mShip Value: %s\033[22m\n", common.FormatPrice(km.ShipPrice))
	res += "\n"
	enrichedItems := common.EnrichedItems(filterItems(km.Victim.EnrichedItems))
	sort.Sort(sort.Reverse(enrichedItems))

	for _, item := range enrichedItems {
		res += fmt.Sprintf("%-60s \033[1m\033[32m%-10d\033[22m\033[39m \033[1m\033[31m%-10d\033[22m\033[39m \033[1m%-10s\033[22m\n", item.ItemName, item.QuantityDropped, item.QuantityDestroyed, common.FormatPrice(item.TotalPrice))
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

func filterItems(items *[]common.EnrichedItem) []common.ItemAggregated {
	enrichedItems := make(map[string]common.ItemAggregated)
	for _, item := range *items {
		if item_, ok := enrichedItems[item.ItemName]; ok {
			if item.QuantityDropped != 0 {
				item_.QuantityDropped += item.QuantityDropped
				item_.TotalPrice += item.Price * float64(item.QuantityDropped)
				enrichedItems[item.ItemName] = item_
			} else {
				item_.QuantityDestroyed += item.QuantityDestroyed
				item_.TotalPrice += item.Price * float64(item.QuantityDestroyed)
				enrichedItems[item.ItemName] = item_
			}
		} else {
			if item.QuantityDropped != 0 {
				enrichedItems[item.ItemName] = common.ItemAggregated{ItemName: item.ItemName, QuantityDropped: item.QuantityDropped, TotalPrice: item.Price * float64(item.QuantityDropped)}
			} else {
				enrichedItems[item.ItemName] = common.ItemAggregated{ItemName: item.ItemName, QuantityDestroyed: item.QuantityDestroyed, TotalPrice: item.Price * float64(item.QuantityDestroyed)}
			}
		}
	}
	res := []common.ItemAggregated{}
	for _, item := range enrichedItems {
		res = append(res, item)
	}
	return res
}

func getKillmailStatus(km *common.EnrichedKMShort) bool {
	return km.Victim.CorporationID == PKID
}

func getDamagePercent(damageDone uint, damageTaken uint) float64 {
	return float64(damageDone) / float64(damageTaken) * 100
}
