package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Pragmatic-Kernel/EveGonline/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var mappings map[uint]string

type KMWithMap struct {
	Killmail    common.Killmail    `json:"killmail"`
	SolarSystem common.SolarSystem `json:"solar_system"`
	Mapping     map[uint]string    `json:"mapping"`
}

func getKMs(db *gorm.DB, w http.ResponseWriter, _ *http.Request) {
	EnrichedKMs := []common.EnrichedKMShort{}
	KMs := []common.Killmail{}
	db.Preload("Attackers").Joins("Victim").Order("killmail_time desc").Find(&KMs)
	for _, km := range KMs {
		mapping := getKMMapping(&km)
		enrichedKM := common.EnrichedKMShort{}
		enrichedKM.ID = km.ID
		enrichedKM.KillmailTime = km.KillmailTime
		enrichedKM.MoonID = km.MoonID
		enrichedKM.WarID = km.WarID
		enrichedKM.Victim = common.EnrichedVictim{Victim: *km.Victim}
		attackers := *km.Attackers
		attacker := filterAttackers(attackers)
		enrichedKM.Attacker = common.EnrichedAttacker{Attacker: attacker}
		enrichKMShort(&enrichedKM, mapping)
		solarSystem := common.GetSolarSystem(db, km.SolarSystemID)
		enrichedKM.SolarSystem = *solarSystem
		EnrichedKMs = append(EnrichedKMs, enrichedKM)
	}
	body, err := json.Marshal(EnrichedKMs)
	if err != nil {
		fmt.Println("ERROR sending KMs")
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(body)
}

func getKM(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	kmIdstr := strings.Split(r.URL.Path, "/")[2]
	kmId, err := strconv.ParseUint(kmIdstr, 10, 64)
	if err != nil {
		fmt.Println("Cannot parse km ID")
		return
	}
	km := common.Killmail{}
	db.Where("id = ?", kmId).Preload("Attackers").Preload("Victim.Items.SubItems").Preload("Attackers").Find(&km)
	solarSystem := common.GetSolarSystem(db, km.SolarSystemID)
	mapping := getKMMapping(&km)
	ekm := common.EnrichedKM{SolarSystem: *solarSystem}
	ekm.Victim = common.EnrichedVictim{Victim: *km.Victim}
	items := []common.EnrichedItem{}
	if km.Victim.Items != nil {
		for _, item_ := range *km.Victim.Items {
			item := common.EnrichedItem{Item: item_}
			if item_.SubItems != nil {
				subitems := []common.EnrichedSubItem{}
				for _, subitem := range *item_.SubItems {
					subitems = append(subitems, common.EnrichedSubItem{SubItem: subitem})
				}
				item.EnrichedSubItems = &subitems
			}
			items = append(items, item)
		}
	}
	ekm.Victim.EnrichedItems = &items
	ekm.KillmailTime = km.KillmailTime
	ekm.MoonID = km.MoonID
	ekm.WarID = km.WarID
	attackers := []common.EnrichedAttacker{}
	for _, attacker := range *km.Attackers {
		attackers = append(attackers, common.EnrichedAttacker{Attacker: attacker})
	}
	ekm.Attackers = &attackers
	enrichKM(&ekm, mapping)

	body, err := json.Marshal(ekm)
	if err != nil {
		fmt.Println("ERROR sending KMs")
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(body)
}

func getKMMapping(km *common.Killmail) map[uint]string {
	globalMapping := mappings
	mapping := make(map[uint]string)
	for _, attacker := range *km.Attackers {
		mapping[attacker.CharacterID] = globalMapping[attacker.CharacterID]
		mapping[attacker.CorporationID] = globalMapping[attacker.CorporationID]
		mapping[attacker.ShipTypeID] = globalMapping[attacker.ShipTypeID]
		mapping[attacker.WeaponTypeID] = globalMapping[attacker.WeaponTypeID]
	}
	mapping[km.Victim.CharacterID] = globalMapping[km.Victim.CharacterID]
	mapping[km.Victim.CorporationID] = globalMapping[km.Victim.CorporationID]
	mapping[km.Victim.ShipTypeID] = globalMapping[km.Victim.ShipTypeID]
	if km.Victim.Items != nil {
		for _, item := range *km.Victim.Items {
			mapping[item.ItemTypeID] = globalMapping[item.ItemTypeID]
			if item.SubItems != nil {
				for _, subitem := range *item.SubItems {
					mapping[subitem.ItemTypeID] = globalMapping[subitem.ItemTypeID]
				}
			}
		}
	}
	return mapping
}

func enrichKMShort(km *common.EnrichedKMShort, mapping map[uint]string) {
	km.Victim.CharName = mapping[km.Victim.CharacterID]
	km.Victim.CorpName = mapping[km.Victim.CorporationID]
	km.Victim.ShipName = mapping[km.Victim.ShipTypeID]
	km.Attacker.CharName = mapping[km.Attacker.CharacterID]
	km.Attacker.CorpName = mapping[km.Attacker.CorporationID]
	km.Attacker.ShipTypeName = mapping[km.Attacker.ShipTypeID]
	km.Attacker.WeaponTypeName = mapping[km.Attacker.WeaponTypeID]
}

func enrichKM(km *common.EnrichedKM, mapping map[uint]string) {
	km.Victim.CharName = mapping[km.Victim.CharacterID]
	km.Victim.CorpName = mapping[km.Victim.CorporationID]
	km.Victim.ShipName = mapping[km.Victim.ShipTypeID]
	attackers := []common.EnrichedAttacker{}
	for _, attacker := range *km.Attackers {
		attacker.CharName = mapping[attacker.CharacterID]
		attacker.CorpName = mapping[attacker.CorporationID]
		attacker.ShipTypeName = mapping[attacker.ShipTypeID]
		attacker.WeaponTypeName = mapping[attacker.WeaponTypeID]
		attackers = append(attackers, attacker)
	}
	km.Attackers = &attackers
	if km.Victim.EnrichedItems != nil {
		enrichedItems := []common.EnrichedItem{}
		for _, item := range *km.Victim.EnrichedItems {
			item.ItemName = mapping[item.ItemTypeID]
			if item.SubItems != nil {
				subitems := []common.EnrichedSubItem{}
				for _, subitem := range *item.EnrichedSubItems {
					subitem.SubItemName = mapping[subitem.ItemTypeID]
					subitems = append(subitems, subitem)
				}
				item.EnrichedSubItems = &subitems
			}
			enrichedItems = append(enrichedItems, item)
		}
		km.Victim.EnrichedItems = &enrichedItems
	}
}

func filterAttackers(attackers []common.Attacker) common.Attacker {
	for _, attacker := range attackers {
		if attacker.FinalBlow {
			return attacker
		}
	}
	return attackers[0]
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	mappings, err = common.GetMappings(db)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/killmails/", func(w http.ResponseWriter, r *http.Request) {
		getKMs(db, w, r)
	})
	mux.HandleFunc("/killmail/", func(w http.ResponseWriter, r *http.Request) {
		getKM(db, w, r)
	})
	mux.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		getImage(db, w, r)
	})
	s := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	s.ListenAndServe()
}
