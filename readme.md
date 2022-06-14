JsonPath
----------------

![Build Status](https://travis-ci.org/oliveagle/jsonpath.svg?branch=master)

Golang实现的JsonPath解析库
- 可以对Json的Value进行读、写
- 可以修改和删除Json的Key

本库基于这个 [库](https://github.com/oliveagle/jsonpath) 进行修改.

感谢 [oliveagle](https://github.com/oliveagle).

快速开始
------------

```bash
go get github.com/denmushi/jsonpath
```

**示例代码**

示例Json
```go
import (
	"encoding/json"
)

var data = {
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

var json_data interface{}
_ = json.Unmarshal([]byte(data), &json_data)
```

读取Json值
```go
import (
    "github.com/denmushi/jsonpath"
)

res, _ := jsonpath.Lookup(json_data, "$.store.book[*].price")
```
`res`是一个`map[string]interface{}`，`key`是解析得到的不含通配符的固定路径，可用于值修改，`value`是该路径对应的值

修改Json值
```go
import (
    "github.com/denmushi/jsonpath"
)

_ = jsonpath.SetToBody(json_data, "$.store.book[0].category", "haha")
```

删除Json的key和value
```go
import (
    "github.com/denmushi/jsonpath"
)

// 根据一个通配路径进行删除
_ = jsonpath.DeleteByKey(json_data, "$.store.book[*].price")

// 根据一个固定路径数组进行删除，采用先标记再删除的方式，不用担心删除过程中json结构会发生变化
_ = jsonpath.DeleteBody(json_data, []string{"$.store.book[0].price","$.store.bicycle"})
```

修改Json的key
```go
import (
    "github.com/denmushi/jsonpath"
)

config := jsonpath.RenamesConfig{
    Config: []jsonpath.RenameConfig{
	    {
	    	From: "$.store.book[*].title",
	        To:   "$.store.new_book[*].new_title",
	    },
	    {
	    	From: "$.expensive",
	    	To:   "$.new_expensive",
	    },
	}
}
_ = jsonpath.Rename(json_data, config)
```

支持提取json模版 "${}"表示模版
```go
import (
"github.com/denmushi/jsonpath"
)

jsonStr := `{
    "url":"${url_}",
    "user":{
		"name": "${name_}",
		"age": 12
	},
    "extra": [
		{
			"e1": "${e1_}",
			"e2": "${e1_}"
		},
		{
			"e3": "${e3_}11",
			"e4": "1${e24_}"
		}
	]
 }`
res, _ := jsonpath.ParseJsonTemplate(jsonStr)
```