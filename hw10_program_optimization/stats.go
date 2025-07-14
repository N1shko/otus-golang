package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type (
	DomainStat map[string]int
	Email      struct {
		Email string `json:"email"`
	}
)

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	suffix := "." + domain

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var tmp Email
		if err := json.Unmarshal(line, &tmp); err != nil {
			return nil, fmt.Errorf("error unmarshalling user: %w", err)
		}

		at := strings.LastIndexByte(tmp.Email, '@')
		if at < 0 || at+1 >= len(tmp.Email) {
			continue
		}

		domainPart := strings.ToLower(tmp.Email[at+1:])
		if strings.HasSuffix(domainPart, suffix) {
			result[domainPart]++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return result, nil
}
