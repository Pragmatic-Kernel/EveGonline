package common

import (
	"fmt"

	"gorm.io/gorm"
)

func GetMappings(db *gorm.DB) (map[uint]string, error) {
	res := make(map[uint]string)
	mappings := []Mapping{}
	db.Find(&mappings)
	for _, mapping := range mappings {
		res[mapping.ID] = mapping.Name
	}
	return res, nil
}

func GetSolarSystem(db *gorm.DB, solarSystemID uint) *SolarSystem {
	solarSystem := SolarSystem{}
	db.Where("id = ?", solarSystemID).Find(&solarSystem)
	return &solarSystem
}

func FormatPrice(price float64) string {
	if price >= 1000000000 {
		price = price / 1000000000
		return fmt.Sprintf("%.2f b ISK", price)
	}
	if price >= 1000000 {
		price = price / 1000000
		return fmt.Sprintf("%.2f m ISK", price)
	}
	if price >= 1000 {
		price = price / 1000
		return fmt.Sprintf("%.2f k ISK", price)
	}
	return fmt.Sprintf("%.2f ISK", price)
}
