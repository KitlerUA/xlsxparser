package xlsxparser

import (
	"fmt"

	"strings"

	"github.com/KitlerUA/xlsxparser/config"
	"github.com/KitlerUA/xlsxparser/policy"
)

const missing = "MISSING"

//Parse - accept records and bindings as matrix
//return slice of policies and warnings
func Parse(records [][]string, bindings [][]string) ([]policy.Policy, []string) {
	var warnings []string
	if len(records) == 1 {
		warnings = append(warnings, fmt.Sprintf("action table is empty"))
	}
	//count of prefix parameters like source and name of policy
	prefixLen := 2
	//current binding position
	currentBinding := 0
	//empty column header error
	emptyHeaderErrors := make(map[int]struct{})
	//empty page errors
	emptyPageErrors := make(map[int]struct{})
	//missing sources in config
	missingSourceErrors := make(map[string]map[int]struct{})
	//missing binding names
	bindingNameErrors := make(map[string]struct{})

	var result []policy.Policy
	for j := prefixLen; j < len(records[0]); j++ {
		//check header of column, if empty - skipp column
		if records[0][j] == "" {
			emptyHeaderErrors[j] = struct{}{}
			continue
		}
		srcPol := make(map[string]*policy.Policy)
		//walk down to collect info about actions and form policy for all role-sources pair for current role
		for i := 1; i < len(records); i++ {
			//split source (more then one source can be in one cell)
			sources := strings.Split(records[i][0], ",")
			for s := range sources {
				//cut spaces
				src := strings.ToLower(strings.TrimSpace(sources[s]))
				if src == "" {
					emptyPageErrors[i+1] = struct{}{}
					continue
				}
				//check if source in config list
				if _, ok := config.Get().PagesNames[src]; !ok {
					if _, ok := missingSourceErrors[src]; !ok {
						missingSourceErrors[src] = make(map[int]struct{})
					}
					missingSourceErrors[src][i+1] = struct{}{}
					continue
				}
				//if record for source doesn't exist - create
				if _, ok := srcPol[src]; !ok {
					var name, description, subject, fileName string
					//take info from table, otherwise - send warning and set fields 'missing value'
					if currentBinding < len(bindings) {
						names := strings.Split(bindings[currentBinding][1], ":")
						name = fmt.Sprintf("pn:%s:%s:%s", strings.ToLower(bindings[currentBinding][0]), strings.ToLower(config.Get().PagesNames[src]), strings.ToLower(names[len(names)-1]))
						description = bindings[currentBinding][2]
						subject = bindings[currentBinding][1]
						fileName = fmt.Sprintf("%s_%s", strings.ToLower(names[len(names)-1]), config.Get().PagesNames[src])

					} else {
						bindingNameErrors[strings.ToLower(records[0][j])] = struct{}{}
						name = fmt.Sprintf("%s:%s", strings.ToLower(records[0][j]), missing)
						description = missing
						subject = missing
						fileName = fmt.Sprintf("%s_%s", strings.ToLower(records[0][j]), missing)
					}
					srcPol[src] = &policy.Policy{
						Name:        name,
						Description: description,
						Subjects:    []string{subject},
						Effect:      "allow",
						Conditions:  policy.Condition{},
						Resources:   []string{fmt.Sprintf("rn:%s", strings.ToLower(config.Get().PagesNames[src]))},
						Actions:     make([]string, 0),
						FileName:    fileName,
					}

				}
				if strings.ToLower(records[i][j]) == strings.ToLower("Yes") {
					srcPol[src].Actions = append(srcPol[src].Actions, records[i][1])
				}
			}
		}
		for _, v := range srcPol {
			result = append(result, *v)
		}
		currentBinding++
	}
	for i := range emptyHeaderErrors {
		warnings = append(warnings, fmt.Sprintf("find empty role-header on %d-th position", i-prefixLen+1))
	}
	for i := range bindingNameErrors {
		warnings = append(warnings, fmt.Sprintf("cannot find binding name for '%s'", i))
	}
	for i := range emptyPageErrors {
		warnings = append(warnings, fmt.Sprintf("found empty page-field on row %d", i))
	}
	for s := range missingSourceErrors {
		for r := range missingSourceErrors[s] {
			warnings = append(warnings, fmt.Sprintf("page '%s' (row %d) isn't in config file: skipped", s, r))
		}
	}
	return result, warnings
}
