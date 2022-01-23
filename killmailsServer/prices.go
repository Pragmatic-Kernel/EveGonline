package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Pragmatic-Kernel/EveGonline/common"
)

func getPrices() (map[uint]float64, error) {
	prices, err := getPricesFromCache()
	if err != nil {
		return nil, fmt.Errorf("unable to get prices from cache file: %w", err)
	}
	if prices != nil {
		pricesMap := getPricesMap(prices)
		return pricesMap, nil
	}
	fmt.Println("Getting file from ESI.")
	prices, err = getPricesFromESI()
	if err != nil {
		return nil, fmt.Errorf("unable to get prices from ESI: %w", err)
	}
	pricesMap := getPricesMap(prices)
	return pricesMap, nil
}

func getPricesFromCache() (*[]common.ItemPrice, error) {
	payload, err := common.GetCache("prices", "market", 86400*3)
	if err != nil {
		if err == common.ErrCacheExpired {
			fmt.Println("File too old, moving.")
			common.MoveCacheFile("prices", "market")
		}
		return nil, fmt.Errorf("unable to get cache file: %w", err)
	}
	prices := []common.ItemPrice{}
	if payload != nil {
		err := json.Unmarshal(payload, &prices)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal cache file: %w", err)
		}
	}
	return &prices, nil
}

func getPricesFromESI() (*[]common.ItemPrice, error) {
	req, err := http.NewRequest("GET", common.EvePricesAPIUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create GET request for prices: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "CharName: Laszlo Bariani")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute GET request for prices: %w", err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read GET request body for prices: %w", err)
	}
	prices := []common.ItemPrice{}
	if payload != nil {
		payload, err = common.SetCache("prices", "market", payload)
		if err != nil {
			return nil, fmt.Errorf("unable to set cache for prices: %w", err)
		}
		err := json.Unmarshal(payload, &prices)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal request body for prices: %w", err)
		}
	}
	return &prices, nil
}

func getPricesMap(itemPrices *[]common.ItemPrice) map[uint]float64 {
	priceMap := make(map[uint]float64)
	for _, itemPrice := range *itemPrices {
		if itemPrice.AveragePrice == 0.0 {
			priceMap[itemPrice.ItemTypeID] = itemPrice.AdjustedPrice
		} else {
			priceMap[itemPrice.ItemTypeID] = itemPrice.AveragePrice
		}
	}
	return priceMap
}

func pricesMapToJson(priceMap map[uint]float64) ([]byte, error) {
	return json.Marshal(priceMap)
}

func getKMPrice(km *common.EnrichedKMShort, priceMap map[uint]float64) *common.EnrichedKMShort {
	price := 0.0
	price += priceMap[km.Victim.ShipTypeID]
	for _, item := range *km.Victim.Items {
		itemPrice := priceMap[item.ItemTypeID]
		priceDropped := itemPrice * float64(item.QuantityDropped)
		priceDestroyed := itemPrice * float64(item.QuantityDestroyed)
		price += priceDropped
		price += priceDestroyed
	}
	km.Price = price
	return km
}
