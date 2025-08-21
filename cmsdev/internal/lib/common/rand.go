// MIT License
//
// (C) Copyright 2019-2025 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
package common

import (
	"fmt"
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

// Return a random string from a list of strings
// Only returns error in the case that the list is empty
func GetRandomStringFromList(stringList []string) (randString string, err error) {
	if len(stringList) > 0 {
		listIndex := IntInRange(0, len(stringList)-1)
		randString = stringList[listIndex]
	} else {
		err = fmt.Errorf("GetRandomStringFromList: Cannot choose a string from an empty list")
	}
	return
}

// Returns a random string from a list of strings except for the one specified
func GetRandomStringFromListExcept(stringList []string, except string) (randString string, err error) {
	var filteredList []string
	if len(stringList) == 0 {
		err = fmt.Errorf("Tenant list is empty")
		return
	}
	for _, s := range stringList {
		if s != except {
			filteredList = append(filteredList, s)
		}
	}
	if len(filteredList) == 0 {
		err = fmt.Errorf("Every string in the list is the excepted string (%s)", except)
		return
	}
	return GetRandomStringFromList(filteredList)
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
