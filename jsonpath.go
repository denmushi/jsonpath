package jsonpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	keyType        = "key"
	idxType        = "idx"
	rangeType      = "range"
	filterType     = "filter"
	scanType       = "scan"
	scanFilterType = "scanFilter"
	rootType       = "root"
)

var (
	errGetFromNullObj = errors.New("get attribute from null object")
	errNotSupported   = errors.New("not supported")
)

func Lookup(obj interface{}, jsonPath string) (map[string]interface{}, error) {
	c, err := Compile(jsonPath)
	if err != nil {
		return nil, err
	}
	return c.Lookup(obj)
}

type step struct {
	op   string
	key  string
	args interface{}
}

func Compile(jsonPath string) (*Compiled, error) {
	tokens, err := tokenize(jsonPath)
	if err != nil {
		return nil, err
	}
	if tokens[0] != "@" && tokens[0] != "$" {
		return nil, fmt.Errorf("$ or @ should in front of path")
	}
	tokens = tokens[1:]
	res := Compiled{
		path:  jsonPath,
		steps: make([]step, len(tokens)),
	}
	for i, t := range tokens {
		op, key, args, err := parseToken(t)
		if err != nil {
			return nil, err
		}
		res.steps[i] = step{op, key, args}
	}
	return &res, nil
}

type Compiled struct {
	path   string
	steps  []step
	result map[string]interface{}
	root   interface{}
}

func (c *Compiled) init(obj interface{}) {
	c.result = make(map[string]interface{})
	if len(c.steps) > 0 {
		c.result["$"] = obj
		c.root = obj
	}
}

func (c *Compiled) String() string {
	return fmt.Sprintf("Compiled lookup: %s", c.path)
}

func (c *Compiled) Lookup(obj interface{}) (map[string]interface{}, error) {
	var err error
	c.init(obj)
	for _, s := range c.steps {
		switch s.op {
		case keyType:
			err = c.getKeys(s)
			if err != nil {
				return nil, err
			}
		case idxType:
			err = c.getIdxes(s)
			if err != nil {
				return nil, err
			}
		case rangeType:
			err = c.getRanges(s)
			if err != nil {
				return nil, err
			}
		case filterType:
			err = c.filter(s)
			if err != nil {
				return nil, err
			}
		case scanFilterType:
			err = c.scanFilter(s)
			if err != nil {
				return nil, err
			}
		case scanType:
			err = c.scans()
			if err != nil {
				return nil, err
			}
		case rootType:
			// todo
		default:
			return nil, errNotSupported
		}
	}
	return c.result, nil
}

func (c *Compiled) scans() error {
	res := make(map[string]interface{})
	for k, v := range c.result {
		obj, err := c.scan(v)
		if err != nil {
			return err
		}
		for oKey, oVal := range obj {
			switch oKey.objType {
			case reflect.Map:
				res[k+"."+oKey.key] = oVal
			case reflect.Slice:
				res[k+"["+fmt.Sprintf("%d", oKey.index)+"]"] = oVal
			default:

			}
		}
	}
	c.result = res
	return nil
}

func (c *Compiled) scan(obj interface{}) (map[objType]interface{}, error) {
	res := make(map[objType]interface{})
	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Slice:
		length := value.Len()
		for i := 0; i < length; i++ {
			ot := objType{
				objType: reflect.Slice,
				index:   i,
			}
			res[ot] = value.Index(i).Interface()
		}
	case reflect.Map:
		it := value.MapRange()
		for it.Next() {
			ot := objType{
				objType: reflect.Map,
				key:     it.Key().String(),
			}
			res[ot] = it.Value().Interface()
		}
	default:
		return nil, errNotSupported
	}
	return res, nil
}

func (c *Compiled) scanFilter(s step) error {
	res := make(map[string]interface{})
	if err := c.scans(); err != nil {
		return err
	}
	for k, v := range c.result {
		obj, err := c.getFiltered(v, s.args.(string))
		if err != nil {
			return err
		}
		for oKey, oVal := range obj {
			switch oKey.objType {
			case reflect.Slice:
				res[k+"["+fmt.Sprintf("%d", oKey.index)+"]"] = oVal
			case reflect.Map:
				res[k+"."+oKey.key] = oVal
			default:
				return errNotSupported
			}

		}
	}
	c.result = res
	return nil
}

func (c *Compiled) filter(s step) error {
	res := make(map[string]interface{})
	if err := c.getKeys(s); err != nil {
		return err
	}
	for k, v := range c.result {
		obj, err := c.getFiltered(v, s.args.(string))
		if err != nil {
			return err
		}
		for oKey, oVal := range obj {
			switch oKey.objType {
			case reflect.Slice:
				res[k+"["+fmt.Sprintf("%d", oKey.index)+"]"] = oVal
			case reflect.Map:
				res[k+"."+oKey.key] = oVal
			default:
				return errNotSupported
			}

		}
	}
	c.result = res
	return nil
}

func (c *Compiled) getRanges(s step) error {
	var err error
	res := make(map[string]interface{})
	if len(s.key) > 0 {
		// no key `$[:1].test`
		err = c.getKeys(s)
		if err != nil {
			return err
		}
	}
	if args, ok := s.args.([2]interface{}); ok {
		for k, v := range c.result {
			obj, err := c.getRange(v, args[0], args[1])
			if err != nil {
				return err
			}
			for oKey, oVal := range obj {
				switch oKey.objType {
				case reflect.Slice:
					res[k+"["+strconv.Itoa(oKey.index)+"]"] = oVal
				case reflect.Map:
					res[k+"."+oKey.key] = oVal
				default:

				}
			}
		}
		c.result = res
	}
	return nil
}

func (c *Compiled) getKeys(s step) error {
	res := make(map[string]interface{})
	for k, obj := range c.result {
		val, err := c.getKey(obj, s.key)
		if err != nil {
			return err
		}
		if val == nil {
			continue
		}
		res[k+"."+s.key] = val
	}
	c.result = res
	return nil
}

func (c *Compiled) getIdxes(s step) error {
	var (
		err error
	)
	if len(s.key) > 0 {
		// no key `$[0].test`
		err = c.getKeys(s)
		if err != nil {
			return err
		}
	}
	if len(s.args.([]int)) > 0 {
		res := make(map[string]interface{})
		for _, x := range s.args.([]int) {
			//fmt.Println("idx ---- ", x)
			for k, v := range c.result {
				tmp, err := c.getIdx(v, x)
				if err != nil {
					return err
				}
				res[k+"["+strconv.Itoa(x)+"]"] = tmp
			}
		}
		c.result = res
	}
	return nil
}

func (c *Compiled) getKey(obj interface{}, key string) (interface{}, error) {
	if reflect.TypeOf(obj) == nil {
		return nil, nil
	}
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Map:
		// if obj came from stdlib json, its highly likely to be a map[string]interface{}
		// in which case we can save having to iterate the map keys to work out if the
		// key exists
		if jsonMap, ok := obj.(map[string]interface{}); ok {
			val, exists := jsonMap[key]
			if !exists {
				return nil, nil
				//return fmt.Errorf("key error: %s not found in object", key)
			}
			return val, nil
		}
		return nil, nil
	//case reflect.Slice:
	//	// slice we should get from all objects in it.
	//	res := []interface{}{}
	//	for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
	//		tmp, _ := getIdx(obj, i)
	//		if v, err := getKey(tmp, key); err == nil {
	//			res = append(res, v)
	//		}
	//	}
	//	//return res, nil
	//	return nil
	default:
		return nil, nil
	}
}

func (c *Compiled) getIdx(obj interface{}, idx int) (interface{}, error) {
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		length := reflect.ValueOf(obj).Len()
		if idx >= 0 {
			if idx >= length {
				return nil, fmt.Errorf("index out of range: len: %v, idx: %v", length, idx)
			}
			return reflect.ValueOf(obj).Index(idx).Interface(), nil
		} else {
			// < 0
			_idx := length + idx
			if _idx < 0 {
				return nil, fmt.Errorf("index out of range: len: %v, idx: %v", length, idx)
			}
			return reflect.ValueOf(obj).Index(_idx).Interface(), nil
		}
	default:
		return nil, fmt.Errorf("object is not Slice")
	}
}

func (c *Compiled) getRange(obj, frm, to interface{}) (map[objType]interface{}, error) {
	res := make(map[objType]interface{})
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		length := reflect.ValueOf(obj).Len()
		_frm := 0
		_to := length
		if frm == nil {
			frm = 0
		}
		if to == nil {
			to = length - 1
		}
		if fv, ok := frm.(int); ok {
			if fv < 0 {
				_frm = length + fv
			} else {
				_frm = fv
			}
		}
		if tv, ok := to.(int); ok {
			if tv < 0 {
				_to = length + tv + 1
			} else {
				_to = tv + 1
			}
		}
		if _frm < 0 || _frm >= length {
			return nil, fmt.Errorf("index [from] out of range: len: %v, from: %v", length, frm)
		}
		if _to < 0 || _to > length {
			return nil, fmt.Errorf("index [to] out of range: len: %v, to: %v", length, to)
		}
		for i := _frm; i < _to; i++ {
			v := reflect.ValueOf(obj).Index(i).Interface()
			ot := objType{
				objType: reflect.Slice,
				index:   i,
			}
			res[ot] = v
		}
		return res, nil
	case reflect.Map:
		// must be *
		if frm != nil || to != nil {
			return nil, fmt.Errorf("get from map error")
		}
		it := reflect.ValueOf(obj).MapRange()
		for it.Next() {
			ot := objType{
				objType: reflect.Map,
				key:     it.Key().String(),
			}
			res[ot] = it.Value().Interface()
		}
		return res, nil
	default:
		return nil, fmt.Errorf("object is not supported")
	}
}

func (c *Compiled) getFiltered(obj interface{}, filter string) (map[objType]interface{}, error) {
	lp, op, rp, err := parseFilter(filter)
	if err != nil {
		return nil, err
	}

	res := make(map[objType]interface{})
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		if op == "=~" {
			// regexp
			pat, err := regFilterCompile(rp)
			if err != nil {
				return nil, err
			}

			for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
				tmp := reflect.ValueOf(obj).Index(i).Interface()
				ok, err := evalRegFilter(tmp, c.root, lp, pat)
				if err != nil {
					return nil, err
				}
				if ok {
					fo := objType{
						objType: reflect.Slice,
						index:   i,
					}
					res[fo] = tmp
				}
			}
		} else {
			for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
				tmp := reflect.ValueOf(obj).Index(i).Interface()
				ok, err := evalFilter(tmp, c.root, lp, op, rp)
				if err != nil {
					return nil, err
				}
				if ok {
					fo := objType{
						objType: reflect.Slice,
						index:   i,
					}
					res[fo] = tmp
				}
			}
		}
		return res, nil
	case reflect.Map:
		if op == "=~" {
			// regexp
			pat, err := regFilterCompile(rp)
			if err != nil {
				return nil, err
			}

			for _, kv := range reflect.ValueOf(obj).MapKeys() {
				tmp := reflect.ValueOf(obj).MapIndex(kv).Interface()
				ok, err := evalRegFilter(tmp, c.root, lp, pat)
				if err != nil {
					return nil, err
				}
				if ok {
					fo := objType{
						objType: reflect.Map,
						key:     kv.String(),
					}
					res[fo] = tmp
				}
			}
		} else {
			for _, kv := range reflect.ValueOf(obj).MapKeys() {
				tmp := reflect.ValueOf(obj).MapIndex(kv).Interface()
				ok, err := evalFilter(tmp, c.root, lp, op, rp)
				if err != nil {
					return nil, err
				}
				if ok {
					fo := objType{
						objType: reflect.Map,
						key:     kv.String(),
					}
					res[fo] = tmp
				}
			}
		}
	default:
		return nil, nil
	}

	return res, nil
}

type objType struct {
	objType reflect.Kind
	index   int
	key     string
}

func tokenize(query string) ([]string, error) {
	tokens := []string{}
	//	token_start := false
	//	token_end := false
	t := ""

	// fmt.Println("-------------------------------------------------- start")
	for idx, x := range query {
		t += string(x)
		// //fmt.Printf("idx: %d, x: %s, token: %s, tokens: %v\n", idx, string(x), token, tokens)
		if idx == 0 {
			if t == "$" || t == "@" {
				tokens = append(tokens, t[:])
				t = ""
				continue
			} else {
				return nil, fmt.Errorf("should start with '$'")
			}
		}
		if t == "." {
			continue
		} else if t == ".." {
			if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, "*")
			}
			t = "."
			continue
		} else {
			// fmt.Println("else: ", string(x), token)
			if strings.Contains(t, "[") {
				// fmt.Println(" contains [ ")
				if x == ']' && !strings.HasSuffix(t, "\\]") {
					if t[0] == '.' {
						tokens = append(tokens, t[1:])
					} else {
						tokens = append(tokens, t[:])
					}
					t = ""
					continue
				}
			} else {
				// fmt.Println(" doesn't contains [ ")
				if x == '.' {
					if t[0] == '.' {
						tokens = append(tokens, t[1:len(t)-1])
					} else {
						tokens = append(tokens, t[:len(t)-1])
					}
					t = "."
					continue
				}
			}
		}
	}
	if len(t) > 0 {
		if t[0] == '.' {
			t = t[1:]
			if t != "*" {
				tokens = append(tokens, t[:])
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, t[:])
			}
		} else {
			if t != "*" {
				tokens = append(tokens, t[:])
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, t[:])
			}
		}
	}
	// fmt.Println("finished tokens: ", tokens)
	// fmt.Println("================================================= done ")
	return tokens, nil
}

/*
 op: "root", "keyType", "idx", "range", "filter", "scan"
*/
func parseToken(token string) (op string, key string, args interface{}, err error) {
	if token == "$" {
		return rootType, "$", nil, nil
	}
	if token == "*" {
		return scanType, "*", nil, nil
	}

	bracketIdx := strings.Index(token, "[")
	if bracketIdx < 0 {
		return keyType, token, nil, nil
	} else {
		key = token[:bracketIdx]
		tail := token[bracketIdx:]
		if len(tail) < 3 {
			err = fmt.Errorf("len(tail) should >=3, %v", tail)
			return
		}
		tail = tail[1 : len(tail)-1]

		//fmt.Println(key, tail)
		if strings.Contains(tail, "?") {
			// filter -------------------------------------------------
			op = filterType
			if strings.HasPrefix(tail, "?(") && strings.HasSuffix(tail, ")") {
				args = strings.Trim(tail[2:len(tail)-1], " ")
			}
			if key == "*" {
				op = scanFilterType
			}
			return
		} else if strings.Contains(tail, ":") {
			// range ----------------------------------------------
			op = rangeType
			tails := strings.Split(tail, ":")
			if len(tails) != 2 {
				err = fmt.Errorf("only support one range(from, to): %v", tails)
				return
			}
			var frm interface{}
			var to interface{}
			if frm, err = strconv.Atoi(strings.Trim(tails[0], " ")); err != nil {
				if strings.Trim(tails[0], " ") == "" {
					err = nil
				}
				frm = nil
			}
			if to, err = strconv.Atoi(strings.Trim(tails[1], " ")); err != nil {
				if strings.Trim(tails[1], " ") == "" {
					err = nil
				}
				to = nil
			}
			args = [2]interface{}{frm, to}
			return
		} else if tail == "*" {
			op = rangeType
			args = [2]interface{}{nil, nil}
			return
		} else {
			// idx ------------------------------------------------
			op = idxType
			res := []int{}
			for _, x := range strings.Split(tail, ",") {
				if i, err := strconv.Atoi(strings.Trim(x, " ")); err == nil {
					res = append(res, i)
				} else {
					return "", "", nil, err
				}
			}
			args = res
		}
	}
	return op, key, args, nil
}

func filterGetFromExplicitPath(obj interface{}, path string) (interface{}, error) {
	steps, err := tokenize(path)
	//fmt.Println("f: steps: ", steps, err)
	//fmt.Println(path, steps)
	if err != nil {
		return nil, err
	}
	if steps[0] != "@" && steps[0] != "$" {
		return nil, fmt.Errorf("$ or @ should in front of path")
	}
	steps = steps[1:]
	xobj := obj
	//fmt.Println("f: xobj", xobj)
	for _, s := range steps {
		op, key, args, err := parseToken(s)
		// key, idx
		switch op {
		case keyType:
			xobj, err = getKey(xobj, key)
			if err != nil {
				return nil, err
			}
		case idxType:
			if len(args.([]int)) != 1 {
				return nil, fmt.Errorf("don't support multiple index in filterType")
			}
			xobj, err = getKey(xobj, key)
			if err != nil {
				return nil, err
			}
			xobj, err = getIdx(xobj, args.([]int)[0])
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("expression don't support in filterType")
		}
	}
	return xobj, nil
}

func getKey(obj interface{}, key string) (interface{}, error) {
	if reflect.TypeOf(obj) == nil {
		return nil, errGetFromNullObj
	}
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Map:
		// if obj came from stdlib json, its highly likely to be a map[string]interface{}
		// in which case we can save having to iterate the map keys to work out if the
		// key exists
		if jsonMap, ok := obj.(map[string]interface{}); ok {
			val, exists := jsonMap[key]
			if !exists {
				return nil, fmt.Errorf("keyType error: %s not found in object", key)
			}
			return val, nil
		}
		for _, kv := range reflect.ValueOf(obj).MapKeys() {
			//fmt.Println(kv.String())
			if kv.String() == key {
				return reflect.ValueOf(obj).MapIndex(kv).Interface(), nil
			}
		}
		return nil, fmt.Errorf("keyType error: %s not found in object", key)
	case reflect.Slice:
		// slice we should get from all objects in it.
		res := []interface{}{}
		for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
			tmp, _ := getIdx(obj, i)
			if v, err := getKey(tmp, key); err == nil {
				res = append(res, v)
			}
		}
		return res, nil
	default:
		return nil, fmt.Errorf("object is not map")
	}
}

func getIdx(obj interface{}, idx int) (interface{}, error) {
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		length := reflect.ValueOf(obj).Len()
		if idx >= 0 {
			if idx >= length {
				return nil, fmt.Errorf("index out of range: len: %v, idx: %v", length, idx)
			}
			return reflect.ValueOf(obj).Index(idx).Interface(), nil
		} else {
			// < 0
			_idx := length + idx
			if _idx < 0 {
				return nil, fmt.Errorf("index out of range: len: %v, idx: %v", length, idx)
			}
			return reflect.ValueOf(obj).Index(_idx).Interface(), nil
		}
	default:
		return nil, fmt.Errorf("object is not Slice")
	}
}

func regFilterCompile(rule string) (*regexp.Regexp, error) {
	runes := []rune(rule)
	if len(runes) <= 2 {
		return nil, errors.New("empty rule")
	}

	if runes[0] != '/' || runes[len(runes)-1] != '/' {
		return nil, errors.New("invalid syntax. should be in `/pattern/` form")
	}
	runes = runes[1 : len(runes)-1]
	return regexp.Compile(string(runes))
}

// @.isbn                 => @.isbn, exists, nil
// @.price < 10           => @.price, <, 10
// @.price <= $.expensive => @.price, <=, $.expensive
// @.author =~ /.*REES/i  => @.author, match, /.*REES/i

func parseFilter(filter string) (lp string, op string, rp string, err error) {
	tmp := ""

	stage := 0
	strEmbrace := false
	for idx, c := range filter {
		switch c {
		case '\'':
			if strEmbrace == false {
				strEmbrace = true
			} else {
				switch stage {
				case 0:
					lp = tmp
				case 1:
					op = tmp
				case 2:
					rp = tmp
				}
				tmp = ""
			}
		case ' ':
			if strEmbrace == true {
				tmp += string(c)
				continue
			}
			switch stage {
			case 0:
				lp = tmp
			case 1:
				op = tmp
			case 2:
				rp = tmp
			}
			tmp = ""

			stage += 1
			if stage > 2 {
				return "", "", "", errors.New(fmt.Sprintf("invalid char at %d: `%c`", idx, c))
			}
		default:
			tmp += string(c)
		}
	}
	if tmp != "" {
		switch stage {
		case 0:
			lp = tmp
			op = "exists"
		case 1:
			op = tmp
		case 2:
			rp = tmp
		}
		tmp = ""
	}
	return lp, op, rp, err
}

func evalRegFilter(obj, root interface{}, lp string, pat *regexp.Regexp) (res bool, err error) {
	if pat == nil {
		return false, errors.New("nil pat")
	}
	lp_v, err := getLpV(obj, root, lp)
	if err != nil {
		return false, err
	}
	switch v := lp_v.(type) {
	case string:
		return pat.MatchString(v), nil
	default:
		return false, errors.New("only string can match with regular expression")
	}
}

func getLpV(obj, root interface{}, lp string) (interface{}, error) {
	var lpV interface{}
	if strings.HasPrefix(lp, "@.") {
		return filterGetFromExplicitPath(obj, lp)
	} else if strings.HasPrefix(lp, "$.") {
		return filterGetFromExplicitPath(root, lp)
	} else {
		lpV = lp
	}
	return lpV, nil
}

func evalFilter(obj, root interface{}, lp, op, rp string) (res bool, err error) {
	lp_v, err := getLpV(obj, root, lp)

	if op == "exists" {
		return lp_v != nil, nil
	} else if op == "=~" {
		return false, fmt.Errorf("not implemented yet")
	} else {
		var rp_v interface{}
		if strings.HasPrefix(rp, "@.") {
			rp_v, err = filterGetFromExplicitPath(obj, rp)
		} else if strings.HasPrefix(rp, "$.") {
			rp_v, err = filterGetFromExplicitPath(root, rp)
		} else {
			rp_v = rp
		}
		//fmt.Printf("lp_v: %v, rp_v: %v\n", lp_v, rp_v)
		return cmpAny(lp_v, rp_v, op)
	}
}

func isNumber(o interface{}) bool {
	switch v := o.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	case string, json.Number:
		_, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		if err == nil {
			return true
		} else {
			return false
		}
	}
	return false
}

func cmpAny(obj1, obj2 interface{}, op string) (bool, error) {
	switch op {
	case "<", "<=", "==", ">=", ">":
	default:
		return false, fmt.Errorf("op should only be <, <=, ==, >= and >")
	}

	var exp string
	if isNumber(obj1) && isNumber(obj2) {
		exp = fmt.Sprintf(`%v %s %v`, obj1, op, obj2)
	} else {
		exp = fmt.Sprintf(`"%v" %s "%v"`, obj1, op, obj2)
	}
	//fmt.Println("exp: ", exp)
	fset := token.NewFileSet()
	res, err := types.Eval(fset, nil, 0, exp)
	if err != nil {
		return false, err
	}
	if res.IsValue() == false || (res.Value.String() != "false" && res.Value.String() != "true") {
		return false, fmt.Errorf("result should only be true or false")
	}
	if res.Value.String() == "true" {
		return true, nil
	}

	return false, nil
}
