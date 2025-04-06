package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(str string) []string {
	str = strings.Replace(str, "\n", " ", -1)
	str = strings.Replace(str, "\t", " ", -1)
	split := strings.Split(str, " ")
	wordsMapped := map[string]int{}
	for _, word := range split {
		if word == "" {
			continue
		}
		if val, ok := wordsMapped[word]; ok {
			wordsMapped[word] = val + 1
		} else {
			wordsMapped[word] = 1
		}
	}

	sliceSorted := make([]string, 0, len(wordsMapped))
	for key := range wordsMapped {
		sliceSorted = append(sliceSorted, key)
	}
	sort.Slice(sliceSorted, func(i, j int) bool {
		if wordsMapped[sliceSorted[i]] == wordsMapped[sliceSorted[j]] {
			return sliceSorted[i] < sliceSorted[j]
		} else {
			return wordsMapped[sliceSorted[i]] > wordsMapped[sliceSorted[j]]
		}
	})
	if len(sliceSorted) < 10 {
		return sliceSorted

	} else {
		return sliceSorted[:10]
	}
}
