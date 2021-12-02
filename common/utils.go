package common

import "gorm.io/gorm"

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
