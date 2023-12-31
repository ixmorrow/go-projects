// Package nutriscore provides utilities for calculating nutritional score and
// Nutri-Score.
// More about-score: https://en.wikipedia.org/wiki/Nutri-Score
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ScoreType int

const (
	Food ScoreType = iota
	Beverage
	Water
	Cheese
)

type NutritionalData struct {
	Energy              EnergyKJ            `json:"energyKj"`
	Sugars              SugarGram           `json:"sugar"`
	SaturatedFattyAcids SaturatedFattyAcids `json:"saturatedFattyAcids"`
	Sodium              SodiumMilligram     `json:"sodiumMg"`
	Fruits              FruitsPercent       `json:"fruitesPercent"`
	Fiber               FiberGram           `json:"fiberGram"`
	Protein             ProteinGram         `json:"proteinGram"`
	IsWater             bool                `json:"isWater"`
	FoodType            ScoreType           `json:"foodType"`
}

var gradeScale = []string{"A", "B", "C", "D", "E"}

var energyLevels = []float64{3350, 3015, 2680, 2345, 2010, 1675, 1340, 1005, 670, 335}
var sugarsLevels = []float64{45, 40, 36, 31, 27, 22.5, 18, 13.5, 9, 4.5}
var saturatedFattyAcidsLevels = []float64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
var sodiumLevels = []float64{900, 810, 720, 630, 540, 450, 360, 270, 180, 90}
var fiberLevels = []float64{4.7, 3.7, 2.8, 1.9, 0.9}
var proteinLevels = []float64{8, 6.4, 4.8, 3.2, 1.6}

var energyLevelsBeverage = []float64{270, 240, 210, 180, 150, 120, 90, 60, 30, 0}
var sugarsLevelsBeverage = []float64{13.5, 12, 10.5, 9, 7.5, 6, 4.5, 3, 1.5, 0}

type NutritionalScore struct {
	Value     int
	Grade     string
	Positive  int
	Negative  int
	ScoreType ScoreType
}

// EnergyKJ represents the energy density in kJ/100g
type EnergyKJ float64

// SugarGram represents amount of sugars in grams/100g
type SugarGram float64

// SaturatedFattyAcids represents amount of saturated fatty acids in grams/100g
type SaturatedFattyAcids float64

// SodiumMilligram represents amount of sodium in mg/100g
type SodiumMilligram float64

// FruitsPercent represents fruits, vegetables, pulses, nuts, and rapeseed, walnut and olive oils as percentage of the total
type FruitsPercent float64

// FibreGram represents amount of fibre in grams/100g
type FiberGram float64

// ProteinGram represents amount of protein in grams/100g
type ProteinGram float64

// EnergyFromKcal converts energy density from kcal to EnergyKJ
func EnergyFromKcal(kcal float64) EnergyKJ {
	return EnergyKJ(kcal * 4.184)
}

// SodiumFromSalt converts salt mg/100g content to sodium content
func SodiumFromSalt(saltMg float64) SodiumMilligram {
	return SodiumMilligram(saltMg / 2.5)
}

// GetPoints returns the nutritional score
func (e EnergyKJ) GetPoints(st ScoreType) int {
	if st == Beverage {
		return getPointsFromRange(float64(e), energyLevelsBeverage)
	}
	return getPointsFromRange(float64(e), energyLevels)
}

// GetPoints returns the nutritional score
func (s SugarGram) GetPoints(st ScoreType) int {
	if st == Beverage {
		return getPointsFromRange(float64(s), sugarsLevelsBeverage)
	}
	return getPointsFromRange(float64(s), sugarsLevels)
}

// GetPoints returns the nutritional score
func (sfa SaturatedFattyAcids) GetPoints(st ScoreType) int {
	return getPointsFromRange(float64(sfa), saturatedFattyAcidsLevels)
}

// GetPoints returns the nutritional score
func (s SodiumMilligram) GetPoints(st ScoreType) int {
	return getPointsFromRange(float64(s), sodiumLevels)
}

// GetPoints returns the nutritional score
func (f FruitsPercent) GetPoints(st ScoreType) int {
	if st == Beverage {
		if f > 80 {
			return 10
		} else if f > 60 {
			return 4
		} else if f > 40 {
			return 2
		}
		return 0
	}
	if f > 80 {
		return 5
	} else if f > 60 {
		return 2
	} else if f > 40 {
		return 1
	}
	return 0
}

// GetPoints returns the nutritional score
func (f FiberGram) GetPoints(st ScoreType) int {
	return getPointsFromRange(float64(f), fiberLevels)
}

// GetPoints returns the nutritional score
func (p ProteinGram) GetPoints(st ScoreType) int {
	return getPointsFromRange(float64(p), proteinLevels)
}

// CalcNutritionalScore calculates the nutritional score for nutritional data n of type st
func CalcNutritionalScore(n NutritionalData) NutritionalScore {
	value := 0
	positive := 0
	negative := 0
	st := n.FoodType
	// Water is always graded A page 30
	if st != Water {
		fruitPoints := n.Fruits.GetPoints(st)
		fibrePoints := n.Fiber.GetPoints(st)
		//negative points are the negative things like calories (it says energy but these are what people are avoiding as these are calories)
		//sugars, saturated fats and sodium
		//positives are fruit points, fiber points and proteins
		negative = n.Energy.GetPoints(st) + n.Sugars.GetPoints(st) + n.SaturatedFattyAcids.GetPoints(st) + n.Sodium.GetPoints(st)
		positive = fruitPoints + fibrePoints + n.Protein.GetPoints(st)

		if st == Cheese {
			// Cheeses always use (negative - positive) page 29
			value = negative - positive
		} else {
			// page 27
			if negative >= 11 && fruitPoints < 5 {
				value = negative - fibrePoints - fruitPoints
			} else {
				value = negative - positive
			}
		}
	}
	return NutritionalScore{
		Value:     value,
		Grade:     n.CalcNutriGrade(value),
		Positive:  positive,
		Negative:  negative,
		ScoreType: st,
	}
}

func (ns NutritionalData) CalcNutriGrade(score int) string {
	if ns.FoodType == Food {
		return gradeScale[getPointsFromRange(float64(score), []float64{18, 10, 2, -1})]
	}
	if ns.FoodType == Water {
		return gradeScale[0]
	}
	return gradeScale[getPointsFromRange(float64(score), []float64{9, 5, 1, -2})]
}

func getPointsFromRange(v float64, levels []float64) int {
	lenLevels := len(levels)
	for i, l := range levels {
		if v > l {
			return lenLevels - i
		}
	}
	return 0
}

func GetNutritionalScore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var nutritionalInfo NutritionalData
	_ = json.NewDecoder(r.Body).Decode(&nutritionalInfo)
	fmt.Println("Nutritional Data Received:", nutritionalInfo)

	nutri_score := CalcNutritionalScore(nutritionalInfo)
	fmt.Printf("Nutritional Score: %d\n", nutri_score.Value)
	fmt.Printf("Nutritional Grade: %s\n", nutri_score.Grade)

	json.NewEncoder(w).Encode(nutri_score)
}
