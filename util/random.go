package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomInt(min, max int64) int64 {
	return min+rand.Int63n(max - min + 1)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i :=0; i < n; i ++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomMoney() int64 {
	return RandomInt(1, 1000)
}

func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)

	return currencies[rand.Intn(n)]
}

func RandomContinent() string {
	continents := []string{"Africa","Antarctica","Asia","Australia","Europe","North America","South America" }
	n := len(continents)

	return continents[rand.Intn(n)]
}

// getRandomOrderStatus returns a random order status from a predefined list
func RandomOrderStatus() string {
	statuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
	return statuses[RandomInt(0, int64(len(statuses)-1))]
}

// util/random.go (add this function)
func RandomProductName() string {
	adjectives := []string{"Premium", "Deluxe", "Basic", "Advanced", "Eco-Friendly", "Smart", "Wireless", "Portable"}
	nouns := []string{"Widget", "Gadget", "Device", "Tool", "Accessory", "System", "Kit", "Bundle"}
	
	adjective := adjectives[RandomInt(0, int64(len(adjectives)-1))]
	noun := nouns[RandomInt(0, int64(len(nouns)-1))]
	number := RandomInt(100, 999)
	
	return fmt.Sprintf("%s %s %d", adjective, noun, number)
}