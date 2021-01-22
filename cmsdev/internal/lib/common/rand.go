/*
Copyright 2019, Cray Inc.  All Rights Reserved.
Author: Torrey Cuthbert <tcuthbert@cray.com>
*/
package common

import (
	"math/rand"
	"time"
)

const alphaLowerChars = "abcdefghijklmnopqrstuvwxyz"
const alphaUpperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numberChars = "0123456789"
const alphaChars = alphaLowerChars + alphaUpperChars
const alnumLowerChars = alphaLowerChars + numberChars
const alnumUpperChars = alphaUpperChars + numberChars
const alnumChars = alphaChars + numberChars

var myRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// These next few functions are just wrappers for the corresponding math/rand
// functions, only using the instance of rand we created and seeded above
func Intn(n int) int {
	return myRand.Intn(n)
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
	for i:= range out {
		j := Intn(numchars)
		out[i] = chars[j]
	}
	return string(out)
}

func AlnumString(length int) string {
	return StringFromChars(length, alnumChars)
}

func AlphaString(length int) string {
	return StringFromChars(length, alphaChars)
}

func AlnumLowerString(length int) string {
	return StringFromChars(length, alnumLowerChars)
}

func AlphaLowerString(length int) string {
	return StringFromChars(length, alphaLowerChars)
}

func AlnumUpperString(length int) string {
	return StringFromChars(length, alnumUpperChars)
}

func AlphaUpperString(length int) string {
	return StringFromChars(length, alphaUpperChars)
}

func NumberString(length int) string {
	return StringFromChars(length, numberChars)
}