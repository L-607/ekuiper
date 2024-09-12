# JQ 函数

jq 函数使用 [jq](https://github.com/itchyny/gojq) 过滤器在运行时查询或转换 JSON 数据。

## JQ

```text
jq(input, query)
```

对输入的 JSON 数据应用 jq 过滤表达式。

**参数：**

- `input`：JSON 字符串、对象或数组。空字符串或 `null` 返回 `nil`。
- `query`：jq 过滤表达式字符串。

**返回值：**

- 无结果：返回 `nil`。
- 单个结果：直接返回该值（字符串/数字/布尔/对象/数组）。
- 多个结果：返回 JSON 数组字符串。

## 示例

### 基本字段访问

```sql
-- 提取字段
jq('{"name":"ekuiper","version":"1.0"}', '.name')
-- 返回: "ekuiper"

-- 嵌套字段访问
jq('{"name":"ekuiper","details":{"version":"1.0","features":["streaming","analytics"]}}', '.details.features[1]')
-- 返回: "analytics"
```

### 数组操作

```sql
-- 数组索引
jq('[{"name":"ekuiper"},{"name":"kuiper"}]', '.[0].name')
-- 返回: "ekuiper"

-- 遍历所有元素
jq('{"name":"ekuiper","version":"1.0"}', '.[]')
-- 返回: "[\"ekuiper\",\"1.0\"]"

-- 数组切片
jq('[1,2,3,4,5]', '.[1:4]')
-- 返回: [2,3,4]

-- 索引越界返回 nil
jq('[1,2]', '.[5]')
-- 返回: nil

-- 数组长度
jq('[1,2,3,4]', 'length')
-- 返回: 4

-- 数组扁平化
jq('[[1,2],[3,[4,5]]]', 'flatten')
-- 返回: [1,2,3,4,5]

-- 数组反转
jq('[1,2,3]', 'reverse')
-- 返回: [3,2,1]

-- 数组排序
jq('[3,1,4,1,5]', 'sort')
-- 返回: [1,1,3,4,5]
```

### 过滤和选择

```sql
-- 使用 select 过滤
jq('[{"name":"ekuiper","version":"1.0"},{"name":"kuiper","version":"2.0"}]', '.[] | select(.version == "2.0") | .name')
-- 返回: "kuiper"

-- map 结合 select 过滤
jq('[1,2,3,4]', 'map(select(.>2))')
-- 返回: [3, 4]

-- 嵌套 select 返回多个结果
jq('[{"tags":["edge","iot"]},{"tags":["edge"]}]', '.[] | .tags[] | select(. == "edge")')
-- 返回: "[\"edge\",\"edge\"]"
```

### 对象操作

```sql
-- 获取所有 key
jq('{"name":"ekuiper","version":"1.0"}', 'keys')
-- 返回: ["name","version"]

-- 获取所有 value
jq('{"name":"ekuiper","version":"1.0"}', '[.[] ]')
-- 返回: ["ekuiper","1.0"]

-- 检查 key 是否存在
jq('{"name":"ekuiper"}', 'has("name")')
-- 返回: true

-- 构造新对象
jq('{"items":[{"a":1},{"a":2}]}', '.items[] | {b: (.a * 2)}')
-- 返回: "[{\"b\":2},{\"b\":4}]"

-- 删除字段
jq('{"name":{"firstname":"Void","lastname":"King"}}', 'del(.name.firstname)')
-- 返回: {"name":{"lastname":"King"}}

-- 添加/修改字段
jq('{"name":"ekuiper"}', '.version = "2.0"')
-- 返回: {"name":"ekuiper","version":"2.0"}
```

### 算术和字符串操作

```sql
-- 数值运算
jq('{"num":3}', '.num * 3')
-- 返回: 9

jq('{"a":10,"b":3}', '.a + .b')
-- 返回: 13

jq('{"a":10,"b":3}', '.a - .b')
-- 返回: 7

-- 字符串拼接
jq('{"str":"hello"}', '.str + " world"')
-- 返回: "hello world"

-- 查找子串索引
jq('"abcb"', 'indices("b")')
-- 返回: [1, 3]

-- 字符串长度
jq('"hello"', 'length')
-- 返回: 5
```

### 条件逻辑

```sql
-- if-then-else
jq('{"age":25}', 'if .age >= 18 then "adult" else "minor" end')
-- 返回: "adult"

-- 比较运算
jq('{"a":5,"b":3}', '.a > .b')
-- 返回: true

-- 逻辑运算
jq('{"a":5,"b":3}', '.a > 3 and .b < 5')
-- 返回: true
```

### 边界情况

```sql
-- 空对象访问字段返回 nil
jq('{}', '.name')
-- 返回: nil

-- 空数组索引返回 nil
jq('[]', '.[0]')
-- 返回: nil

-- 空字符串输入返回 nil
jq('', '.a')
-- 返回: nil
```
