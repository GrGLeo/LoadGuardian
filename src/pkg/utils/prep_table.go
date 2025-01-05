package utils

import (
	"errors"
	"strconv"
	"strings"
)

func GenerateTable(rows [][]string, baseLength []int) string {
  var table string
  var tableLength int
  for i := range rows {
    row, _ := GenerateRow(rows[i], baseLength)
    // Write header
    table += row
    // Write interrow for header
    if i == 0 {
      tableLength = len(row)
    }
    interRow := GenerateInterRow(tableLength)
    table += interRow
  }
  top := GenerateInterRow(tableLength)
  table = top + table
  return table
}

func GenerateRow(row []string, length []int) (string, error) {
  if len(row) != len(length) {
    return "", errors.New("Lists do not contain the same number of items")
  }
  formattedRow := "|"
  pad := " |"
  for i := 0; i < len(row); i++ {
    str := row[i]
    strLength := length[i]
    PadString(&str, strLength)
    str += pad
    formattedRow += str
  }
  formattedRow += "\n"
  return formattedRow, nil
}


func GenerateInterRow(length int) string {
  row := ""
  row += strings.Repeat("-", length - 1)
  row += "\n"
  return row
}


func PadString(str *string, length int)  {
  missingLength := length - len(*str)
  if missingLength > 0 {
    *str += strings.Repeat(" ", missingLength)
  }
}

func ConvertFloatToValue(fl float64, val string) string {
  str := strconv.FormatFloat(fl, 'f', 2, 64)
  str += " " + val
  return str
}

func GetBaseLength(header []string) []int {
  baseLenght := []int{}
  for i := range header {
    baseLenght = append(baseLenght, len(header[i]))
  }
  return baseLenght
}

func UpdateBaseLength(base, row *[]int) error {
  if len(*base) != len(*row) {
    return errors.New("Lists do not containe the same number of items")
  }
  for i := range *base {
    if (*row)[i] > (*base)[i] {
      (*base)[i] = (*row)[i]
    }
  } 
  return nil
}
