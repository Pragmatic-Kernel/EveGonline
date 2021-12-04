package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Pragmatic-Kernel/EveGoNline/common"
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
	KMWithMaps := []KMWithMap{}
	KMs := []common.Killmail{}
	db.Preload("Attackers").Joins("Victim").Find(&KMs)
	for _, km := range KMs {
		mapping := getKMMapping(&km)
		KmWithMap := KMWithMap{}
		KmWithMap.Killmail = km
		KmWithMap.Mapping = mapping
		solarSystem := common.GetSolarSystem(db, km.SolarSystemID)
		KmWithMap.SolarSystem = *solarSystem
		KMWithMaps = append(KMWithMaps, KmWithMap)
	}
	body, err := json.Marshal(KMWithMaps)
	if err != nil {
		fmt.Println("ERROR sending KMs")
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(body)
}

func getKM(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	kmIdstr := strings.Split(r.URL.Path, "/")[2]
	fmt.Println(kmIdstr)
	kmId, err := strconv.ParseUint(kmIdstr, 10, 64)
	if err != nil {
		fmt.Println("Cannot parse km ID")
		return
	}
	km := common.Killmail{}
	db.Where("id = ?", kmId).Preload("Attackers").Preload("Victim.Items.SubItems").Preload("Attackers").Find(&km)
	solarSystem := common.GetSolarSystem(db, km.SolarSystemID)
	mapping := getKMMapping(&km)
	s := KMWithMap{Killmail: km, Mapping: mapping, SolarSystem: *solarSystem}
	body, err := json.Marshal(s)
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
		mapping[attacker.ShipTypeID] = globalMapping[attacker.ShipTypeID]
		mapping[attacker.WeaponTypeID] = globalMapping[attacker.WeaponTypeID]
	}
	mapping[km.Victim.CharacterID] = globalMapping[km.Victim.CharacterID]
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
	fmt.Println(mapping)
	return mapping
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
