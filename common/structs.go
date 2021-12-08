package common

import (
	"time"

	"gorm.io/gorm"
)

type Mapping struct {
	gorm.Model `json:"-"`
	ID         uint   `json:"id"`
	Category   string `json:"category"`
	Name       string `json:"name"`
}

type PreToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type Token struct {
	gorm.Model
	AccessToken  string
	RefreshToken string
	Exp          uint
	CorpID       uint
	CharID       uint
}

type Payload struct {
	Scp   []string `json:"scp"`
	Jti   string   `json:"jti"`
	Kid   string   `json:"kid"`
	Sub   string   `json:"sub"`
	Azp   string   `json:"azp"`
	Name  string   `json:"name"`
	Owner string   `json:"owner"`
	Exp   uint     `json:"exp"`
	Iss   string   `json:"iss"`
}

type Killmail struct {
	gorm.Model    `json:"-"`
	Attackers     *[]Attacker `gorm:"constraint:OnDelete:CASCADE" json:"attackers"`
	ID            uint        `json:"killmail_id"`
	Hash          string      `json:"killmail_hash"`
	KillmailTime  time.Time   `json:"killmail_time"`
	MoonID        uint        `json:"moon_id"`
	SolarSystemID uint        `json:"solar_system_id"`
	Victim        *Victim     `gorm:"constraint:OnDelete:CASCADE" json:"victim"`
	WarID         uint        `json:"war_id"`
}

type Attacker struct {
	gorm.Model     `json:"-"`
	KillmailID     uint    `json:"-"`
	AllianceID     uint    `json:"alliance_id"`
	CharacterID    uint    `json:"character_id"`
	CorporationID  uint    `json:"corporation_id"`
	DamageDone     uint    `json:"damage_done"`
	FactionID      uint    `json:"faction_id"`
	FinalBlow      bool    `json:"final_blow"`
	SecurityStatus float64 `json:"security_status"`
	ShipTypeID     uint    `json:"ship_type_id"`
	WeaponTypeID   uint    `json:"weapon_type_id"`
}

type Victim struct {
	gorm.Model    `json:"-"`
	KillmailID    uint      `json:"-"`
	AllianceID    uint      `json:"alliance_id"`
	CharacterID   uint      `json:"character_id"`
	CorporationID uint      `json:"corporation_id"`
	DamageTaken   uint      `json:"damage_taken"`
	FactionID     uint      `json:"faction_id"`
	Items         *[]Item   `gorm:"constraint:OnDelete:CASCADE" json:"items"`
	Position      *Position `gorm:"constraint:OnDelete:CASCADE" json:"position"`
	ShipTypeID    uint      `json:"ship_type_id"`
}

type Item struct {
	gorm.Model        `json:"-"`
	VictimID          uint       `json:"-"`
	Flag              uint       `json:"flag"`
	ItemTypeID        uint       `json:"item_type_id"`
	SubItems          *[]SubItem `gorm:"constraint:OnDelete:CASCADE" json:"items"`
	QuantityDestroyed uint       `json:"quantity_destroyed"`
	QuantityDropped   uint       `json:"quantity_dropped"`
	Singleton         uint       `json:"singleton"`
}

type SubItem struct {
	gorm.Model        `json:"-"`
	ItemID            uint `json:"-"`
	Flag              uint `json:"flag"`
	ItemTypeID        uint `json:"item_type_id"`
	QuantityDestroyed uint `json:"quantity_destroyed"`
	QuantityDropped   uint `json:"quantity_dropped"`
	Singleton         uint `json:"singleton"`
}

type Position struct {
	gorm.Model `json:"-"`
	VictimID   uint    `json:"-"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Z          float64 `json:"z"`
}

type SolarSystem struct {
	gorm.Model     `json:"-"`
	ID             uint    `json:"solar_system_id:"`
	RegionID       uint    `json:"region_id:"`
	SecurityStatus float64 `json:"security_status"`
	Name           string  `json:"name"`
}

type Asset struct {
	gorm.Model
	Etag string
	Size uint
}

type EnrichedKMShort struct {
	Victim       EnrichedVictim   `json:"victim"`
	Attacker     EnrichedAttacker `json:"attacker"`
	SolarSystem  SolarSystem      `json:"solar_system"`
	ID           uint             `json:"killmail_id"`
	KillmailTime time.Time        `json:"killmail_time"`
	MoonID       uint             `json:"moon_id"`
	WarID        uint             `json:"war_id"`
}

type EnrichedKM struct {
	Victim       EnrichedVictim      `json:"victim"`
	Attackers    *[]EnrichedAttacker `json:"attackers"`
	SolarSystem  SolarSystem         `json:"solar_system"`
	ID           uint                `json:"killmail_id"`
	KillmailTime time.Time           `json:"killmail_time"`
	MoonID       uint                `json:"moon_id"`
	WarID        uint                `json:"war_id"`
}

type EnrichedVictim struct {
	Victim
	CharName       string          `json:"character_name"`
	CharPortrait   string          `json:"character_portrait"`
	CorpName       string          `json:"corporation_name"`
	CorpLogo       string          `json:"corporation_logo"`
	ShipTypeName   string          `json:"ship_type_name"`
	EnrichedItems  *[]EnrichedItem `json:"items"`
	ShipTypeIcon   string          `json:"ship_type_icon"`
	ShipTypeRender string          `json:"ship_type_render"`
}

type EnrichedAttacker struct {
	Attacker
	CharName       string `json:"character_name"`
	CharPortrait   string `json:"character_portrait"`
	CorpName       string `json:"corporation_name"`
	CorpLogo       string `json:"corporation_logo"`
	ShipTypeName   string `json:"ship_type_name"`
	ShipTypeIcon   string `json:"ship_type_icon"`
	WeaponTypeName string `json:"weapon_type_name"`
	WeaponTypeIcon string `json:"weapon_type_icon"`
}

type EnrichedItem struct {
	Item
	EnrichedSubItems *[]EnrichedSubItem `json:"items"`
	ItemName         string             `json:"item_name"`
	ItemIcon         string             `json:"item_icon"`
}

type EnrichedSubItem struct {
	SubItem
	SubItemName string `json:"item_name"`
	SubItemIcon string `json:"item_icon"`
}
