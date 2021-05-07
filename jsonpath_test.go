package jsonpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	jsonData   interface{}
	jsonDataV2 interface{}
	jsonDataV3 interface{}
)

func init() {
	data := `
{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10
}
`
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.UseNumber()
	_ = decoder.Decode(&jsonData)
	dataV2 := `{
    "fields": {
        "人力评估": {
            "name": "manpower",
            "value": 2
        },
        "任务执行人": {
            "name": "executor",
            "value":[
                 {
                    "id": "ou_debc524b2d8cb187704df652b43d29de"
                 }
            ]},
        "任务描述": {
            "name": "description",
            "value":"多渠道收集用户反馈"
        },
        "链接 URL": {
            "name": "url",
            "value":{
                "text": " 多渠道反馈收集表格 ",
                "link": "http://bitable.feishu.cn"
            }
         },
        "对应 OKR": {
            "name": "okr",
            "value":[
                "recqwIwhc6",
                "recOuEJMvN"
            ]},
        "截止日期": {
            "name": "deadline",
            "value": 1609516800000
        },
        "是否完成": {
            "name": "completed",
            "value": true
        },
        "状态": {
            "name": "status",
            "value":"已结束"
        },
        "相关部门": {
            "name": "departments",
            "value": [
                "销售",
                "客服"
            ]
        }
    }
}`
	decoder = json.NewDecoder(strings.NewReader(dataV2))
	decoder.UseNumber()
	_ = decoder.Decode(&jsonDataV2)
	dataV3 := `{
  "data": {
    "records": [
      {
        "record_id": "123",
        "fields": {
          "1": "haha",
          "2": false,
          "3": ["1","2"],
          "4": [
            {
              "id": "ou_xxx",
              "name": "haha",
              "@type": "person"
            },
            {
              "id": "ou_yyy",
              "name": "hhhh"
            }
          ]
        }
      }
    ]
  }
}`
	decoder = json.NewDecoder(strings.NewReader(dataV3))
	decoder.UseNumber()
	_ = decoder.Decode(&jsonDataV3)
}

func Test_jsonpath_JsonPathLookup_1(t *testing.T) {

	t.Run("list express", func(t *testing.T) {
		res, err := Lookup(jsonDataV3, "$.data.records[*].fields.*[?(@.@type == person)].id")
		want := map[string]interface{}{
			"$.data.records[0].fields.4[0].id": "ou_xxx",
		}
		assert.Nil(t, err)
		assert.Equal(t, res, want)
	})
	t.Run("key from root", func(t *testing.T) {
		resV2, err := Lookup(jsonData, "$.expensive")
		assert.Nil(t, err)
		assert.Equal(t, resV2, map[string]interface{}{
			"$.expensive": json.Number("10"),
		})
	})
	t.Run("single index", func(t *testing.T) {
		res, err := Lookup(jsonData, "$.store.book[0].price")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[0].price": json.Number("8.95"),
		})
	})

	t.Run("negative single index", func(t *testing.T) {
		res, err := Lookup(jsonData, "$.store.book[-1].isbn")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[-1].isbn": "0-395-19395-8",
		})
	})

	t.Run("multiple index", func(t *testing.T) {
		res, err := Lookup(jsonData, "$.store.book[0,1].price")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[0].price": json.Number("8.95"),
			"$.store.book[1].price": json.Number("12.99"),
		})
	})

	t.Run("multiple index", func(t *testing.T) {
		res, err := Lookup(jsonData, "$.store.book[0,1].title")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[0].title": "Sayings of the Century",
			"$.store.book[1].title": "Sword of Honour",
		})
	})

	t.Run("full array", func(t *testing.T) {
		res, err := Lookup(jsonData, "$.store.book[0:].price")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[0].price": json.Number("8.95"),
			"$.store.book[1].price": json.Number("12.99"),
			"$.store.book[2].price": json.Number("8.99"),
			"$.store.book[3].price": json.Number("22.99"),
		})
	})

	t.Run("range", func(t *testing.T) {
		res, err := Lookup(jsonData, "$.store.book[0:1].price")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[0].price": json.Number("8.95"),
			"$.store.book[1].price": json.Number("12.99"),
		})
	})

	t.Run("range", func(t *testing.T) {
		res, err := Lookup(jsonData, "$.store.book[0:1].title")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[0].title": "Sayings of the Century",
			"$.store.book[1].title": "Sword of Honour",
		})
	})
}

func Test_jsonpath_JsonPathLookup_filter(t *testing.T) {
	t.Run("filter", func(t *testing.T) {
		res, err := Lookup(jsonData, "$.store.book[?(@.isbn)].isbn")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[2].isbn": "0-553-21311-3",
			"$.store.book[3].isbn": "0-395-19395-8",
		})

		res, err = Lookup(jsonData, "$.store.book[?(@.price > 10)].title")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[1].title": "Sword of Honour",
			"$.store.book[3].title": "The Lord of the Rings",
		})

		res, err = Lookup(jsonData, "$.store.book[?(@.price > 10)]")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[1]": map[string]interface{}{
				"category": "fiction",
				"author":   "Evelyn Waugh",
				"title":    "Sword of Honour",
				"price":    json.Number("12.99"),
			},
			"$.store.book[3]": map[string]interface{}{
				"category": "fiction",
				"author":   "J. R. R. Tolkien",
				"title":    "The Lord of the Rings",
				"isbn":     "0-395-19395-8",
				"price":    json.Number("22.99"),
			},
		})

		res, err = Lookup(jsonData, "$.store.book[?(@.price > $.expensive)].price")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[1].price": json.Number("12.99"),
			"$.store.book[3].price": json.Number("22.99"),
		})

		res, err = Lookup(jsonData, "$.store.book[?(@.price < $.expensive)].price")
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.store.book[0].price": json.Number("8.95"),
			"$.store.book[2].price": json.Number("8.99"),
		})
	})
	t.Run("map filter", func(t *testing.T) {
		res, err := Lookup(jsonDataV2, `$.fields[?(@.name == executor)].value[*].id`)
		assert.Nil(t, err)
		assert.Equal(t, res, map[string]interface{}{
			"$.fields.任务执行人.value[0].id": "ou_debc524b2d8cb187704df652b43d29de",
		})
	})
}

func Test_jsonpath_authors_of_all_books(t *testing.T) {
	query := "$.store.book[*].author"
	res, err := Lookup(jsonData, query)
	assert.Nil(t, err)
	assert.Equal(t, res, map[string]interface{}{
		"$.store.book[0].author": "Nigel Rees",
		"$.store.book[1].author": "Evelyn Waugh",
		"$.store.book[2].author": "Herman Melville",
		"$.store.book[3].author": "J. R. R. Tolkien",
	})
}

var tokenCases = []map[string]interface{}{
	{
		"query":  "$..author",
		"tokens": []string{"$", "*", "author"},
	},
	{
		"query":  "$.store.*",
		"tokens": []string{"$", "store", "*"},
	},
	{
		"query":  "$.store..price",
		"tokens": []string{"$", "store", "*", "price"},
	},
	{
		"query":  "$.store.book[*].author",
		"tokens": []string{"$", "store", "book[*]", "author"},
	},
	{
		"query":  "$..book[2]",
		"tokens": []string{"$", "*", "book[2]"},
	},
	{
		"query":  "$..book[(@.length-1)]",
		"tokens": []string{"$", "*", "book[(@.length-1)]"},
	},
	{
		"query":  "$..book[0,1]",
		"tokens": []string{"$", "*", "book[0,1]"},
	},
	{
		"query":  "$..book[:2]",
		"tokens": []string{"$", "*", "book[:2]"},
	},
	{
		"query":  "$..book[?(@.isbn)]",
		"tokens": []string{"$", "*", "book[?(@.isbn)]"},
	},
	{
		"query":  "$.store.book[?(@.price < 10)]",
		"tokens": []string{"$", "store", "book[?(@.price < 10)]"},
	},
	{
		"query":  "$..book[?(@.price <= $.expensive)]",
		"tokens": []string{"$", "*", "book[?(@.price <= $.expensive)]"},
	},
	{
		"query":  "$..book[?(@.author =~ /.*REES/i)]",
		"tokens": []string{"$", "*", "book[?(@.author =~ /.*REES/i)]"},
	},
	{
		"query":  "$..book[?(@.author =~ /.*REES\\]/i)]",
		"tokens": []string{"$", "*", "book[?(@.author =~ /.*REES\\]/i)]"},
	},
	{
		"query":  "$..*",
		"tokens": []string{"$", "*"},
	},
	{
		"query":  "$....author",
		"tokens": []string{"$", "*", "author"},
	},
}

func TestJsonpathTokenize(t *testing.T) {
	t.Run("tokenize", func(t *testing.T) {
		for idx, tokenCase := range tokenCases {
			t.Logf("idx[%d], tokenCase: %v", idx, tokenCase)
			query := tokenCase["query"].(string)
			expectedTokens := tokenCase["tokens"].([]string)
			tokens, err := tokenize(query)
			assert.Nil(t, err)
			assert.Equal(t, tokens, expectedTokens)
		}
	})
}

var parseTokenCases = []map[string]interface{}{
	{
		"token": "$",
		"op":    "root",
		"key":   "$",
		"args":  nil,
	},
	{
		"token": "store",
		"op":    "key",
		"key":   "store",
		"args":  nil,
	},

	// idx --------------------------------------
	{
		"token": "book[2]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{2},
	},
	{
		"token": "book[-1]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{-1},
	},
	{
		"token": "book[0,1]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{0, 1},
	},
	{
		"token": "[0]",
		"op":    "idx",
		"key":   "",
		"args":  []int{0},
	},

	// range ------------------------------------
	{
		"token": "book[1:-1]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{1, -1},
	},
	{
		"token": "book[*]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{nil, nil},
	},
	{
		"token": "book[:2]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{nil, 2},
	},
	{
		"token": "book[-2:]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{-2, nil},
	},

	// filter --------------------------------
	{
		"token": "book[?( @.isbn      )]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.isbn",
	},
	{
		"token": "book[?(@.price < 10)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.price < 10",
	},
	{
		"token": "book[?(@.price <= $.expensive)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.price <= $.expensive",
	},
	{
		"token": "book[?(@.author =~ /.*REES/i)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.author =~ /.*REES/i",
	},
	{
		"token": "*",
		"op":    "scan",
		"key":   "*",
		"args":  nil,
	},
}

func TestJsonpathParseToken(t *testing.T) {
	t.Run("parse token", func(t *testing.T) {
		for idx, tokenCase := range parseTokenCases {
			t.Logf("[%d] - tokenCase: %v", idx, tokenCase)
			token := tokenCase["token"].(string)
			expOp := tokenCase["op"].(string)
			expKey := tokenCase["key"].(string)
			expArgs := tokenCase["args"]

			op, key, args, err := parseToken(token)
			assert.Nil(t, err)
			assert.Equal(t, op, expOp)
			assert.Equal(t, key, expKey)
			assert.Equal(t, args, expArgs)
		}
	})
}

func TestJsonpathGetKey(t *testing.T) {
	obj := map[string]interface{}{
		"key": 1,
	}
	res, err := getKey(obj, "key")
	fmt.Println(err, res)
	if err != nil {
		t.Errorf("failed to get key: %v", err)
		return
	}
	if res.(int) != 1 {
		t.Errorf("key value is not 1: %v", res)
		return
	}

	res, err = getKey(obj, "hah")
	fmt.Println(err, res)
	if err == nil {
		t.Errorf("key error not raised")
		return
	}
	if res != nil {
		t.Errorf("key error should return nil res: %v", res)
		return
	}

	obj2 := 1
	res, err = getKey(obj2, "key")
	fmt.Println(err, res)
	if err == nil {

		t.Errorf("object is not map error not raised")
		return
	}
	obj3 := map[string]string{"key": "hah"}
	res, err = getKey(obj3, "key")
	if res_v, ok := res.(string); ok != true || res_v != "hah" {
		fmt.Println(err, res)
		t.Errorf("map[string]string support failed")
	}

	obj4 := []map[string]interface{}{
		{
			"a": 1,
		},
		{
			"a": 2,
		},
	}
	res, err = getKey(obj4, "a")
	fmt.Println(err, res)
}

func TestJsonpathGetIdx(t *testing.T) {
	c := &Compiled{}
	obj := []interface{}{1, 2, 3, 4}
	res, err := c.getIdx(obj, 0)
	assert.Nil(t, err)
	assert.Equal(t, res, 1)

	res, err = c.getIdx(obj, 2)
	assert.Nil(t, err)
	assert.Equal(t, res, 3)
	res, err = c.getIdx(obj, 4)
	assert.NotNil(t, err)

	res, err = c.getIdx(obj, -1)
	assert.Nil(t, err)
	assert.Equal(t, res, 4)

	res, err = c.getIdx(obj, -4)
	assert.Nil(t, err)
	assert.Equal(t, res, 1)

	res, err = c.getIdx(obj, -5)
	assert.NotNil(t, err)

	obj1 := 1
	res, err = c.getIdx(obj1, 1)
	assert.NotNil(t, err)
}

func TestJsonpathGetRange(t *testing.T) {
	type (
		testCase struct {
			obj    interface{}
			frm    interface{}
			to     interface{}
			expRes map[objType]interface{}
		}
	)
	cases := []testCase{
		{
			obj: []int{1, 2, 3, 4, 5},
			frm: 0,
			to:  2,
			expRes: map[objType]interface{}{
				{
					objType: reflect.Slice,
					index:   0,
				}: 1,
				{
					objType: reflect.Slice,
					index:   1,
				}: 2,
				{
					objType: reflect.Slice,
					index:   2,
				}: 3,
			},
		},
		{
			obj: []int{1, 2, 3, 4, 5},
			frm: 3,
			to:  -1,
			expRes: map[objType]interface{}{
				{
					objType: reflect.Slice,
					index:   3,
				}: 4,
				{
					objType: reflect.Slice,
					index:   4,
				}: 5,
			},
		},
		{
			obj: []int{1, 2, 3, 4, 5},
			frm: nil,
			to:  2,
			expRes: map[objType]interface{}{
				{
					objType: reflect.Slice,
					index:   0,
				}: 1,
				{
					objType: reflect.Slice,
					index:   1,
				}: 2,
				{
					objType: reflect.Slice,
					index:   2,
				}: 3,
			},
		},
		{
			obj: []int{1, 2, 3, 4, 5},
			frm: nil,
			to:  nil,
			expRes: map[objType]interface{}{
				{
					objType: reflect.Slice,
					index:   0,
				}: 1,
				{
					objType: reflect.Slice,
					index:   1,
				}: 2,
				{
					objType: reflect.Slice,
					index:   2,
				}: 3,
				{
					objType: reflect.Slice,
					index:   3,
				}: 4,
				{
					objType: reflect.Slice,
					index:   4,
				}: 5,
			},
		},
		{
			obj: []int{1, 2, 3, 4, 5},
			frm: -2,
			to:  nil,
			expRes: map[objType]interface{}{
				{
					objType: reflect.Slice,
					index:   3,
				}: 4,
				{
					objType: reflect.Slice,
					index:   4,
				}: 5,
			},
		},
		{
			obj: map[string]interface{}{
				"a": "a1",
				"b": "b1",
				"c": "c1",
			},
			frm: nil,
			to:  nil,
			expRes: map[objType]interface{}{
				{
					objType: reflect.Map,
					key:     "a",
				}: "a1",
				{
					objType: reflect.Map,
					key:     "b",
				}: "b1",
				{
					objType: reflect.Map,
					key:     "c",
				}: "c1",
			},
		},
	}
	c := &Compiled{}
	for _, oneCase := range cases {
		res, err := c.getRange(oneCase.obj, oneCase.frm, oneCase.to)
		assert.Nil(t, err)
		assert.Equal(t, res, oneCase.expRes)
	}

	obj2 := 2
	_, err := c.getRange(obj2, 0, 1)
	assert.NotNil(t, err)
}

var testCaseParseFilter = []map[string]interface{}{
	// 0
	{
		"filter":  "@.isbn",
		"exp_lp":  "@.isbn",
		"exp_op":  "exists",
		"exp_rp":  "",
		"exp_err": nil,
	},
	// 1
	{
		"filter":  "@.price < 10",
		"exp_lp":  "@.price",
		"exp_op":  "<",
		"exp_rp":  "10",
		"exp_err": nil,
	},
	// 2
	{
		"filter":  "@.price <= $.expensive",
		"exp_lp":  "@.price",
		"exp_op":  "<=",
		"exp_rp":  "$.expensive",
		"exp_err": nil,
	},
	// 3
	{
		"filter":  "@.author =~ /.*REES/i",
		"exp_lp":  "@.author",
		"exp_op":  "=~",
		"exp_rp":  "/.*REES/i",
		"exp_err": nil,
	},

	// 4
	{
		"filter": "@.author == 'Nigel Rees'",
		"exp_lp": "@.author",
		"exp_op": "==",
		"exp_rp": "Nigel Rees",
	},
}

func TestJsonpathParseFilter(t *testing.T) {
	for _, testCase := range testCaseParseFilter {
		lp, op, rp, _ := parseFilter(testCase["filter"].(string))
		t.Log(testCase)
		t.Logf("lp: %v, op: %v, rp: %v", lp, op, rp)
		assert.Equal(t, lp, testCase["exp_lp"])
		assert.Equal(t, op, testCase["exp_op"])
		assert.Equal(t, rp, testCase["exp_rp"])
	}
}

var testCaseFilterGetFromExplicitPath = []map[string]interface{}{
	// 0
	{
		// 0 {"a": 1}
		"obj":      map[string]interface{}{"a": 1},
		"query":    "$.a",
		"expected": 1,
	},
	{
		// 1 {"a":{"b":1}}
		"obj":      map[string]interface{}{"a": map[string]interface{}{"b": 1}},
		"query":    "$.a.b",
		"expected": 1,
	},
	{
		// 2 {"a": {"b":1, "c":2}}
		"obj":      map[string]interface{}{"a": map[string]interface{}{"b": 1, "c": 2}},
		"query":    "$.a.c",
		"expected": 2,
	},
	{
		// 3 {"a": {"b":1}, "b": 2}
		"obj":      map[string]interface{}{"a": map[string]interface{}{"b": 1}, "b": 2},
		"query":    "$.a.b",
		"expected": 1,
	},
	{
		// 4 {"a": {"b":1}, "b": 2}
		"obj":      map[string]interface{}{"a": map[string]interface{}{"b": 1}, "b": 2},
		"query":    "$.b",
		"expected": 2,
	},
	{
		// 5 {'a': ['b',1]}
		"obj":      map[string]interface{}{"a": []interface{}{"b", 1}},
		"query":    "$.a[0]",
		"expected": "b",
	},
}

func Test_jsonpath_filter_get_from_explicit_path(t *testing.T) {
	for _, testCase := range testCaseFilterGetFromExplicitPath {
		obj := testCase["obj"]
		query := testCase["query"].(string)
		expected := testCase["expected"]

		res, err := filterGetFromExplicitPath(obj, query)
		assert.Nil(t, err)
		assert.Equal(t, res, expected)
	}
}

var testCaseEvalFilter = []map[string]interface{}{
	// 0
	{
		"obj":  map[string]interface{}{"a": 1},
		"root": map[string]interface{}{},
		"lp":   "@.a",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 1
	{
		"obj":  map[string]interface{}{"a": 1},
		"root": map[string]interface{}{},
		"lp":   "@.b",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},
	// 2
	{
		"obj":  map[string]interface{}{"a": 1},
		"root": map[string]interface{}{"a": 1},
		"lp":   "$.a",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 3
	{
		"obj":  map[string]interface{}{"a": 1},
		"root": map[string]interface{}{"a": 1},
		"lp":   "$.b",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},
	// 4
	{
		"obj":  map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2}},
		"root": map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2}},
		"lp":   "$.b.c",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 5
	{
		"obj":  map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2}},
		"root": map[string]interface{}{},
		"lp":   "$.b.a",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},

	// 6
	{
		"obj":  map[string]interface{}{"a": 3},
		"root": map[string]interface{}{"a": 3},
		"lp":   "$.a",
		"op":   ">",
		"rp":   "1",
		"exp":  true,
	},
}

func TestJsonpathEvalFilter(t *testing.T) {
	for idx, tcase := range testCaseEvalFilter[1:] {
		fmt.Println("------------------------------")
		obj := tcase["obj"].(map[string]interface{})
		root := tcase["root"].(map[string]interface{})
		lp := tcase["lp"].(string)
		op := tcase["op"].(string)
		rp := tcase["rp"].(string)
		exp := tcase["exp"].(bool)
		t.Logf("idx: %v, lp: %v, op: %v, rp: %v, exp: %v", idx, lp, op, rp, exp)
		got, err := evalFilter(obj, root, lp, op, rp)
		assert.Nil(t, err)
		assert.Equal(t, got, exp)

	}
}

var (
	ifc1 interface{} = "haha"
	ifc2 interface{} = "ha ha"
)
var testCaseCmpAny = []map[string]interface{}{
	{
		"obj1": 1,
		"obj2": 1,
		"op":   "==",
		"exp":  true,
		"err":  nil,
	},
	{
		"obj1": 1,
		"obj2": 2,
		"op":   "==",
		"exp":  false,
		"err":  nil,
	},
	{
		"obj1": 1.1,
		"obj2": 2.0,
		"op":   "<",
		"exp":  true,
		"err":  nil,
	},
	{
		"obj1": "1",
		"obj2": "2.0",
		"op":   "<",
		"exp":  true,
		"err":  nil,
	},
	{
		"obj1": "1",
		"obj2": "2.0",
		"op":   ">",
		"exp":  false,
		"err":  nil,
	},
	{
		"obj1": 1,
		"obj2": 2,
		"op":   "=~",
		"exp":  false,
		"err":  errors.New("op should only be <, <=, ==, >= and >"),
	},
	{
		"obj1": ifc1,
		"obj2": ifc1,
		"op":   "==",
		"exp":  true,
		"err":  nil,
	},
	{
		"obj1": ifc2,
		"obj2": ifc2,
		"op":   "==",
		"exp":  true,
		"err":  nil,
	},
	{
		"obj1": 20,
		"obj2": "100",
		"op":   ">",
		"exp":  false,
		"err":  nil,
	},
}

func TestJsonpathCmpAny(t *testing.T) {
	for idx, testCase := range testCaseCmpAny {
		t.Logf("idx: %v, %v %v %v, exp: %v", idx, testCase["obj1"], testCase["op"], testCase["obj2"], testCase["exp"])
		res, err := cmpAny(testCase["obj1"], testCase["obj2"], testCase["op"].(string))
		exp := testCase["exp"].(bool)
		expErr := testCase["err"]
		assert.Equal(t, err, expErr)
		assert.Equal(t, res, exp)
	}
}

func TestJsonpathNullInTheMiddle(t *testing.T) {
	data := `{
		"head_commit": null,
		"test": {
			"author": {
				"username": "Jack"
			}
		}
	}
	`
	type (
		caseType struct {
			data     string
			jsonPath string
			expRes   map[string]interface{}
		}
	)
	cases := []caseType{
		{
			data:     data,
			jsonPath: "$.test[*]",
			expRes: map[string]interface{}{
				"$.test.author": map[string]interface{}{
					"username": "Jack",
				},
			},
		},
		{
			data:     data,
			jsonPath: "$..author.username",
			expRes: map[string]interface{}{
				"$.test.author.username": "Jack",
			},
		},
	}

	var (
		j   interface{}
		err error
	)
	err = json.Unmarshal([]byte(data), &j)
	assert.Nil(t, err)
	for _, oneCase := range cases {
		err = json.Unmarshal([]byte(oneCase.data), &j)
		assert.Nil(t, err)
		res, err := Lookup(j, oneCase.jsonPath)
		assert.Nil(t, err)
		assert.Equal(t, res, oneCase.expRes)
	}
}

//func BenchmarkJsonPathLookup_0(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		JsonPathLookup(jsonData, "$.expensive")
//	}
//}
//
//func BenchmarkJsonPathLookup_1(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		JsonPathLookup(jsonData, "$.store.book[0].price")
//	}
//}
//
//func TestReg(t *testing.T) {
//	r := regexp.MustCompile(`(?U).*REES`)
//	t.Log(r)
//	t.Log(r.Match([]byte(`Nigel Rees`)))
//
//	res, err := JsonPathLookup(jsonData, "$.store.book[?(@.author =~ /(?i).*REES/ )].author")
//	t.Log(err, res)
//
//	author := res.([]interface{})[0].(string)
//	t.Log(author)
//	if author != "Nigel Rees" {
//		t.Fatal("should be `Nigel Rees` but got: ", author)
//	}
//}

var tcases_reg_op = []struct {
	Line string
	Exp  string
	Err  bool
}{
	{``, ``, true},
	{`xxx`, ``, true},
	{`/xxx`, ``, true},
	{`xxx/`, ``, true},
	{`'/xxx/'`, ``, true},
	{`"/xxx/"`, ``, true},
	{`/xxx/`, `xxx`, false},
	{`/π/`, `π`, false},
}

func TestRegOp(t *testing.T) {
	for idx, tcase := range tcases_reg_op {
		fmt.Println("idx: ", idx, "tcase: ", tcase)
		res, err := regFilterCompile(tcase.Line)
		if tcase.Err == true {
			if err == nil {
				t.Fatal("expect err but got nil")
			}
		} else {
			if res == nil || res.String() != tcase.Exp {
				t.Fatal("different. res:", res)
			}
		}
	}
}

func TestJsonpathRootnodeIsArray(t *testing.T) {
	data := `[{
   "test": 12.34
}, {
	"test": 13.34
}, {
	"test": 14.34
}]
`

	var j interface{}
	err := json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Lookup(j, "$[0].test")
	assert.Nil(t, err)
	assert.Equal(t, res, map[string]interface{}{
		"$[0].test": 12.34,
	})
}

func TestJsonpathRootnodeIsArrayRange(t *testing.T) {
	data := `[{
   "test": 12.34
}, {
	"test": 13.34
}, {
	"test": 14.34
}]
`

	var j interface{}

	err := json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Lookup(j, "$[:1].test")
	assert.Nil(t, err)
	assert.Equal(t, res, map[string]interface{}{
		"$[0].test": 12.34,
		"$[1].test": 13.34,
	})
}

func TestJsonpathRootnodeIsNestedArray(t *testing.T) {
	data := `[ [ {"test":1.1}, {"test":2.1} ], [ {"test":3.1}, {"test":4.1} ] ]`

	var j interface{}

	err := json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Lookup(j, "$[0].[0].test")
	assert.Nil(t, err)
	assert.Equal(t, res, map[string]interface{}{
		"$[0][0].test": 1.1,
	})
}

func TestJsonpathRootnodeIsNestedArrayRange(t *testing.T) {
	data := `[ [ {"test":1.1}, {"test":2.1} ], [ {"test":3.1}, {"test":4.1} ] ]`

	var j interface{}

	err := json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Lookup(j, "$[:1].[0].test")
	assert.Nil(t, err)
	assert.Equal(t, res, map[string]interface{}{
		"$[0][0].test": 1.1,
		"$[1][0].test": 3.1,
	})
}
