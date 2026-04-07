package cli

import (
	"fmt"
	"strings"

	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdTags(git *gitutil.Repo) {
	fmt.Println(output.Cyan("Available versions:"))
	fmt.Println()

	tagsOut, err := git.Tags()
	if err != nil || tagsOut == "" {
		output.Warn("No tags yet.")
		return
	}

	for _, t := range strings.Split(tagsOut, "\n") {
		if t == "" {
			continue
		}
		d := git.TagDate(t)
		m := git.TagMessage(t)
		fmt.Printf("  %-12s %s  %s\n", t, d, m)
	}
}
