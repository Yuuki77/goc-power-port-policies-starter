package src

import (
	"strings"
)

type CommitCounter struct {
	CommiterCounterMap map[string]int
	ReviewerCounterMap map[string]int
}

func (c *CommitCounter) AddCommitCount(author string) {
	c.CommiterCounterMap[author]++
}

func (c *CommitCounter) AddReviewCount(text string) {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Reviewed-by:") {
			removedLine := strings.Replace(line, "Reviewed-by: ", "", 1)
			c.ReviewerCounterMap[removedLine]++
		}
	}
}
