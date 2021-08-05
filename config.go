package jsonpath

import (
	"errors"
	"strings"
)

type RenamesConfig struct {
	Config []RenameConfig `json:"config"`
}

type RenameConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type renameConfigParse struct {
	Len       int      `json:"len"` // FromParse和ToParse包括$,Len是扣掉$之后的长度
	FromParse []string `json:"from_parse"`
	ToParse   []string `json:"to_parse"`
}

func (r renameConfigParse) getFromByIndex(index int) (string, error) {
	if index >= r.Len {
		return "", errors.New("out of index")
	}
	return r.FromParse[index+1], nil
}

func (r renameConfigParse) getToByIndex(index int) (string, error) {
	if index >= r.Len {
		return "", errors.New("out of index")
	}
	return r.ToParse[index+1], nil
}

// RenameConfigParse保存的数组第一个是带$的
// 所以index是0的时候，表示要构造的是FromParse[0]+FromParse[1]
func (r renameConfigParse) buildPath(index int) (string, string, bool) {
	if index >= r.Len {
		return "", "", false
	}
	from := strings.Join(r.FromParse[:index+2], ".")
	to := strings.Join(r.ToParse[:index+2], ".")
	return from, to, true
}

func (r renameConfigParse) renameFrom(index int) error {
	if index >= r.Len {
		return errors.New("out of index")
	}
	r.FromParse[index+1] = r.ToParse[index+1]
	return nil
}

func (r RenamesConfig) parseConfig() ([]renameConfigParse, int) {
	result := make([]renameConfigParse, 0, len(r.Config))
	maxLen := -1
	for _, each := range r.Config {
		parse := renameConfigParse{}
		parse.FromParse = strings.Split(each.From, ".")
		parse.ToParse = strings.Split(each.To, ".")
		parse.Len = len(parse.FromParse) - 1
		if parse.Len > maxLen {
			maxLen = parse.Len
		}
		result = append(result, parse)
	}

	return result, maxLen
}
