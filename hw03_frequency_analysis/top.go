package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(str string) []string {
	split := strings.Fields(str)
	wordsMapped := map[string]int{}
	for _, word := range split {
		wordsMapped[word] += 1
	}

	sliceSorted := make([]string, 0, len(wordsMapped))
	for key := range wordsMapped {
		sliceSorted = append(sliceSorted, key)
	}
	sort.Slice(sliceSorted, func(i, j int) bool {
		if wordsMapped[sliceSorted[i]] == wordsMapped[sliceSorted[j]] {
			return sliceSorted[i] < sliceSorted[j]
		}
		return wordsMapped[sliceSorted[i]] > wordsMapped[sliceSorted[j]]
	})
	if len(sliceSorted) < 10 {
		return sliceSorted
	}
	return sliceSorted[:10]
}
