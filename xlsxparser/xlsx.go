package xlsxparser

import (
	"fmt"
	"strings"

	"github.com/tealeg/xlsx"

	"github.com/KitlerUA/xlsxparser/config"
)

//XLSX - read data from .xlsx file and
//return map with key=sheet.Name and matrix of values
//read until first empty row
func XLSX(fileName string) (map[string][][]string, map[string][][]string, string, error) {
	warn := ""
	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		return make(map[string][][]string), make(map[string][][]string), warn, err
	}
	res := make(map[string][][]string)
	bindings := make(map[string][][]string)
	for _, sheet := range xlFile.Sheets {
		//skip empty sheet
		if len(sheet.Rows) < 1 {
			continue
		}
		//search for "page", "name" and "roles" in first row
		page, pageFound, name, nameFound, rolesStart, rolesStartFound, rolesEnd, rolesEndFound := searchHeaders(sheet.Rows[0].Cells)
		//check if all headers was found
		if !(pageFound && nameFound && rolesStartFound && rolesEndFound) {
			warn += fmt.Sprintf("<b>%s</b>: cannot find <i>%s</i>, <i>%s</i> or bounds for roles<br>", sheet.Name, config.Get().Page, config.Get().Name)
			continue
		}
		//position of the end of table
		var tableEndRow int
		res[sheet.Name], tableEndRow, err = parseActionTable(sheet.Name, sheet.Rows, page, name, rolesStart, rolesEnd)
		if err != nil {
			return make(map[string][][]string), make(map[string][][]string), warn, err
		}
		//try to find subjects table
		bindings[sheet.Name] = parseBindingsTable(sheet.Rows, tableEndRow)
	}
	return res, bindings, warn, nil
}

func isRowEmpty(row *xlsx.Row) bool {
	if len(row.Cells) == 0 {
		return true
	}
	for _, r := range row.Cells {
		if r.String() != "" {
			return false
		}
	}
	return true
}

func isPartRowEmpty(row *xlsx.Row, a, b int) bool {
	if len(row.Cells) == 0 {
		return true
	}
	for i := a; i < b && i < len(row.Cells); i++ {
		if row.Cells[i].String() != "" {
			return false
		}
	}
	return true
}

func searchHeaders(cells []*xlsx.Cell) (int, bool, int, bool, int, bool, int, bool) {
	var (
		//index of Page column
		page      int
		pageFound = false
		//index of Name column
		name      int
		nameFound = false
		//indices for roles_start and roles_end
		rolesStart      int
		rolesStartFound = false
		rolesEnd        int
		rolesEndFound   = false
	)
	for j, cell := range cells {
		switch strings.ToLower(cell.String()) {
		case strings.ToLower(config.Get().Page):
			page = j
			pageFound = true
		case strings.ToLower(config.Get().Name):
			name = j
			nameFound = true
		case strings.ToLower(config.Get().RolesBegin):
			rolesStart = j + 1
			rolesStartFound = true
		case strings.ToLower(config.Get().RolesEnd):
			rolesEnd = j - 1
			rolesEndFound = true
		}
	}
	return page, pageFound, name, nameFound, rolesStart, rolesStartFound, rolesEnd, rolesEndFound
}

func parseActionTable(sheetName string, rows []*xlsx.Row, page, name, rolesStart, rolesEnd int) ([][]string, int, error) {
	var tableEndRow int
	res := make([][]string, 0)
	for i, row := range rows {
		//first empty row mean the end of the table
		if isRowEmpty(row) {
			tableEndRow = i
			break
		}
		//add new row
		res = append(res, []string{})
		//insert Page
		res[i] = append(res[i], row.Cells[page].String())
		//insert Name
		res[i] = append(res[i], row.Cells[name].String())
		//insert Roles
		for j := rolesStart; j <= rolesEnd; j++ {
			if rolesEnd >= len(row.Cells) {
				return make([][]string, 0), 0, fmt.Errorf("sheet=%s: find empty tail of row %d<br>Please, fix action's table", sheetName, i)
			}
			res[i] = append(res[i], row.Cells[j].String())
		}
	}
	return res, tableEndRow, nil
}

func parseBindingsTable(rows []*xlsx.Row, tableEndRow int) [][]string {
	bindings := make([][]string, 0)
	for i := tableEndRow + 1; i < len(rows); i++ {
		for j := 0; j < len(rows[i].Cells)-2; j++ {
			cell := rows[i].Cells[j].String()
			cell1 := rows[i].Cells[j+1].String()
			cell2 := rows[i].Cells[j+2].String()
			if strings.ToLower(cell) == strings.ToLower(config.Get().Type) &&
				strings.ToLower(cell1) == strings.ToLower(config.Get().TechGroupName) &&
				strings.ToLower(cell2) == strings.ToLower(config.Get().DisplayName) {
				for r := i + 1; r < len(rows); r++ {
					if isPartRowEmpty(rows[r], j, j+2) || len(rows[r].Cells)-1 < j+2 {
						break
					}
					bindings = append(bindings, []string{rows[r].Cells[j].String(), rows[r].Cells[j+1].String(), rows[r].Cells[j+2].String()})
				}
			}
		}

	}
	return bindings
}
