package csync

import (
	"encoding/csv"
	"os"
)

func getPartners() map[string]partner {
	partnerFile := "./partnerFile.csv"
	partners := make(map[string]partner)

	f, err := os.Open(partnerFile)
	check(err)
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	check(err)

	for _, line := range lines {
		scope := []string{}
		for i := 3; i < len(line); i++ {
			scope = append(scope, line[i])
		}
		// p := partner{line[0], line[1], line[2], scope}
		// partners[line[0]] = p
		partners[line[0]] = partner{line[0], line[1], line[2], scope}
	}
	return partners
}
