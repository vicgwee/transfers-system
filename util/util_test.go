package util

import (
	"math/big"
	"testing"
)

func TestStringToAmount(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		wantErr bool
		want    *big.Rat
	}{
		{name: "Valid amount", val: "123.45", want: big.NewRat(12345, 100)},
		{name: "Invalid amount", val: "invalid", wantErr: true},
		{name: "Valid Rat, Invalid amount", val: "10/1", wantErr: true},
		{name: "NaN amount", val: "NaN", wantErr: true},
		{name: "Inf amount", val: "Inf", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringToAmount(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.Cmp(tt.want) != 0 {
				t.Errorf("StringToAmount() = %v, expected integer value", got)
			}
		})
	}
}

func TestAmountToString(t *testing.T) {
	tests := []struct {
		name string
		val  big.Rat
		want string
	}{
		{"Whole number", *big.NewRat(10, 1), "10.00000"},
		{"Decimal number", *big.NewRat(10, 3), "3.33333"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AmountToString(tt.val); got != tt.want {
				t.Errorf("AmountToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
