/*
Copyright 2019-2021 Hewlett Packard Enterprise Development LP
*/
package common

import (
	"math/rand"
	"time"
)

const AlphaLowerChars = "abcdefghijklmnopqrstuvwxyz"
const AlphaUpperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const NumberChars = "0123456789"
const PunctuationChars = ".,:;<>/?[]{}\\|=+-_)(*&^%$#@!~`'\""
const WhitespaceChars = " \t"
const AlphaChars = AlphaLowerChars + AlphaUpperChars
const AlnumLowerChars = AlphaLowerChars + NumberChars
const AlnumUpperChars = AlphaUpperChars + NumberChars
const AlnumChars = AlphaChars + NumberChars
const TextChars = PunctuationChars + WhitespaceChars + AlnumChars

var myRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// These next few functions are just wrappers for the corresponding math/rand
// functions, only using the instance of rand we created and seeded above
func Intn(n int) int {
	return myRand.Intn(n)
}

// Returns an integer in [min, max]
func IntInRange(min, max int) int {
	if min > max || min < 0 {
		return 0
	}
	return Intn(max+1-min) + min
}

func Float32() float32 {
	return myRand.Float32()
}

func Float64() float64 {
	return myRand.Float64()
}

// Generate a random string of the specified length from the specified characters
func StringFromChars(length int, chars string) string {
	out := make([]byte, length)
	numchars := len(chars)
	for i := range out {
		j := Intn(numchars)
		out[i] = chars[j]
	}
	return string(out)
}

func AlnumString(length int) string {
	return StringFromChars(length, AlnumChars)
}

func AlphaString(length int) string {
	return StringFromChars(length, AlphaChars)
}

func AlnumLowerString(length int) string {
	return StringFromChars(length, AlnumLowerChars)
}

func AlphaLowerString(length int) string {
	return StringFromChars(length, AlphaLowerChars)
}

func AlnumUpperString(length int) string {
	return StringFromChars(length, AlnumUpperChars)
}

func AlphaUpperString(length int) string {
	return StringFromChars(length, AlphaUpperChars)
}

func NumberString(length int) string {
	return StringFromChars(length, NumberChars)
}

func TextString(length int) string {
	return StringFromChars(length, TextChars)
}

func TextStringWithNewlines(length int) string {
	return StringFromChars(length, TextChars+"\n\r")
}
