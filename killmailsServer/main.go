package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pragmatic-Kernel/EveGonline/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var lock sync.RWMutex
var mappings map[uint]string

type KMWithMap struct {
	Killmail    common.Killmail    `json:"killmail"`
	SolarSystem common.SolarSystem `json:"solar_system"`
}

func getKMs(db *gorm.DB, w http.ResponseWriter, _ *http.Request) {
	EnrichedKMs := []common.EnrichedKMShort{}
	KMs := []common.Killmail{}
	priceMap, _ := getPrices()
	db.Preload("Attackers").Preload("Victim.Items.SubItems").Order("killmail_time desc").Find(&KMs)
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
		getKMPrice(&enrichedKM, priceMap)
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
	priceMap, _ := getPrices()
	kmIdstr := strings.Split(r.URL.Path, "/")[2]
	kmId, err := strconv.ParseUint(kmIdstr, 10, 64)
	if err != nil {
		fmt.Println("Cannot parse km ID")
		return
	}
	km := common.Killmail{}
	db.Where("id = ?", kmId).Preload("Attackers").Preload("Victim.Items.SubItems").Preload("Attackers").Find(&km)
	if km.ID == 0 {
		w.WriteHeader(404)
		return
	}
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
	kmshort := getKMPrice(&common.EnrichedKMShort{Victim: ekm.Victim}, priceMap)
	ekm.Price = kmshort.Price

	body, err := json.Marshal(ekm)
	if err != nil {
		fmt.Println("ERROR sending KMs")
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(body)
}

func getKMMapping(km *common.Killmail) map[uint]string {
	lock.RLock()
	globalMapping := mappings
	lock.RUnlock()
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
	km.Victim.CharacterName = mapping[km.Victim.CharacterID]
	km.Victim.CharacterPortrait = getImageURLfromIDTypeSize(km.Victim.CharacterID, "characters", 64)
	km.Victim.CorporationName = mapping[km.Victim.CorporationID]
	km.Victim.CorporationLogo = getImageURLfromIDTypeSize(km.Victim.CorporationID, "corporations", 64)
	km.Victim.ShipTypeName = mapping[km.Victim.ShipTypeID]
	km.Victim.ShipTypeIcon = getImageURLfromIDTypeSize(km.Victim.ShipTypeID, "icons", 64)
	km.Attacker.CharacterName = mapping[km.Attacker.CharacterID]
	km.Attacker.CharacterPortrait = getImageURLfromIDTypeSize(km.Attacker.CharacterID, "characters", 64)
	km.Attacker.CorporationName = mapping[km.Attacker.CorporationID]
	km.Attacker.CorporationLogo = getImageURLfromIDTypeSize(km.Attacker.CorporationID, "corporations", 64)
	km.Attacker.ShipTypeName = mapping[km.Attacker.ShipTypeID]
	km.Attacker.ShipTypeIcon = getImageURLfromIDTypeSize(km.Attacker.ShipTypeID, "icons", 64)
	km.Attacker.WeaponTypeName = mapping[km.Attacker.WeaponTypeID]
	km.Attacker.WeaponTypeIcon = getImageURLfromIDTypeSize(km.Attacker.WeaponTypeID, "icons", 64)
}

func enrichKM(km *common.EnrichedKM, mapping map[uint]string) {
	km.Victim.CharacterName = mapping[km.Victim.CharacterID]
	km.Victim.CharacterPortrait = getImageURLfromIDTypeSize(km.Victim.CharacterID, "characters", 64)
	km.Victim.CorporationName = mapping[km.Victim.CorporationID]
	km.Victim.CorporationLogo = getImageURLfromIDTypeSize(km.Victim.CorporationID, "corporations", 64)
	km.Victim.ShipTypeName = mapping[km.Victim.ShipTypeID]
	km.Victim.ShipTypeIcon = getImageURLfromIDTypeSize(km.Victim.ShipTypeID, "icons", 64)
	km.Victim.ShipTypeRender = getImageURLfromIDTypeSize(km.Victim.ShipTypeID, "renders", 128)
	attackers := []common.EnrichedAttacker{}
	for _, attacker := range *km.Attackers {
		attacker.CharacterName = mapping[attacker.CharacterID]
		attacker.CharacterPortrait = getImageURLfromIDTypeSize(attacker.CharacterID, "characters", 64)
		attacker.CorporationName = mapping[attacker.CorporationID]
		attacker.CorporationLogo = getImageURLfromIDTypeSize(attacker.CorporationID, "corporations", 64)
		attacker.ShipTypeName = mapping[attacker.ShipTypeID]
		attacker.ShipTypeIcon = getImageURLfromIDTypeSize(attacker.ShipTypeID, "icons", 64)
		attacker.WeaponTypeName = mapping[attacker.WeaponTypeID]
		attacker.WeaponTypeIcon = getImageURLfromIDTypeSize(attacker.WeaponTypeID, "icons", 64)
		attackers = append(attackers, attacker)
	}
	km.Attackers = &attackers
	if km.Victim.EnrichedItems != nil {
		enrichedItems := []common.EnrichedItem{}
		for _, item := range *km.Victim.EnrichedItems {
			item.ItemName = mapping[item.ItemTypeID]
			item.ItemIcon = getImageURLfromIDTypeSize(item.ItemTypeID, "icons", 64)
			if item.SubItems != nil {
				subitems := []common.EnrichedSubItem{}
				for _, subitem := range *item.EnrichedSubItems {
					subitem.ItemName = mapping[subitem.ItemTypeID]
					subitem.ItemIcon = getImageURLfromIDTypeSize(subitem.ItemTypeID, "icons", 64)
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

func getImageURLfromIDTypeSize(ID uint, Type string, size uint) string {
	switch Type {
	case "renders":
		return "/images/renders/" + fmt.Sprintf("%d", ID) + "/render?size=" + fmt.Sprintf("%d", size)
	case "icons":
		return "/images/types/" + fmt.Sprintf("%d", ID) + "/icon?size=" + fmt.Sprintf("%d", size)
	case "characters":
		return "/images/characters/" + fmt.Sprintf("%d", ID) + "/portrait?size=" + fmt.Sprintf("%d", size)
	case "corporations":
		return "/images/corporations/" + fmt.Sprintf("%d", ID) + "/logo?size=" + fmt.Sprintf("%d", size)
	}
	return ""
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
	go func() {
		for {
			ticker := time.NewTicker(15 * time.Minute)
			<-ticker.C
			lock.Lock()
			mappings, err = common.GetMappings(db)
			if err != nil {
				panic(err)
			}
			lock.Unlock()
		}
	}()
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
