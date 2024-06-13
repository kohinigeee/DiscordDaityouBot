package lib

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

func GeneRandomID() string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	i := random.Intn(1e8)
	return fmt.Sprintf("%08d", i)
}

func MapKyes[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func ParsePlaceHolders(str string) map[string]string {
	tokens := strings.Split(str, "\n")

	m := make(map[string]string)
	for _, token := range tokens {
		words := strings.Split(token, "=")
		if len(words) < 2 {
			continue
		}
		key := words[0]
		value := words[1]
		if value == "" {
			continue
		}
		m[key] = value
	}

	return m
}
