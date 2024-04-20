package util

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
)

func StringToAmount(val string) (big.Rat, error) {
	r := big.Rat{}

	// First try parsing as float, because fractions can be scanned into big.Rat
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Println("error parsing value as float:", err)
		return r, NewInvalidAmountError(val)
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return r, NewInvalidAmountError(val)
	}

	// Scan as big.Rat to preserve precision
	_, err = fmt.Sscan(val, &r)
	if err != nil {
		log.Println("error scanning value:", err)
		return r, NewInvalidAmountError(val)
	}
	return r, nil
}

func AmountToString(val big.Rat) string {
	return val.FloatString(5)
}
