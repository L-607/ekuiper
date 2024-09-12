# JQ Function

The jq function applies a [jq](https://github.com/itchyny/gojq) filter to query or transform JSON data at runtime.

## JQ

```text
jq(input, query)
```

Apply a jq filter expression to the input JSON data.

**Arguments:**

- `input`: JSON string, object, or array. Empty string or `null` returns `nil`.
- `query`: jq filter expression string.

**Returns:**

- No result: returns `nil`.
- Single result: returns the value directly (string/number/bool/object/array).
- Multiple results: returns a JSON array string.

## Examples

### Basic Field Access

```sql
-- Pick a field
jq('{"name":"ekuiper","version":"1.0"}', '.name')
-- Returns: "ekuiper"

-- Nested field access
jq('{"name":"ekuiper","details":{"version":"1.0","features":["streaming","analytics"]}}', '.details.features[1]')
-- Returns: "analytics"
```

### Array Operations

```sql
-- Array indexing
jq('[{"name":"ekuiper"},{"name":"kuiper"}]', '.[0].name')
-- Returns: "ekuiper"

-- Iterate all elements
jq('{"name":"ekuiper","version":"1.0"}', '.[]')
-- Returns: "[\"ekuiper\",\"1.0\"]"

-- Array slicing
jq('[1,2,3,4,5]', '.[1:4]')
-- Returns: [2,3,4]

-- Index out of range returns nil
jq('[1,2]', '.[5]')
-- Returns: nil

-- Array length
jq('[1,2,3,4]', 'length')
-- Returns: 4

-- Flatten nested arrays
jq('[[1,2],[3,[4,5]]]', 'flatten')
-- Returns: [1,2,3,4,5]

-- Reverse array
jq('[1,2,3]', 'reverse')
-- Returns: [3,2,1]

-- Sort array
jq('[3,1,4,1,5]', 'sort')
-- Returns: [1,1,3,4,5]
```

### Filtering and Selection

```sql
-- Filter with select
jq('[{"name":"ekuiper","version":"1.0"},{"name":"kuiper","version":"2.0"}]', '.[] | select(.version == "2.0") | .name')
-- Returns: "kuiper"

-- Map with select filter
jq('[1,2,3,4]', 'map(select(.>2))')
-- Returns: [3, 4]

-- Nested select with multiple results
jq('[{"tags":["edge","iot"]},{"tags":["edge"]}]', '.[] | .tags[] | select(. == "edge")')
-- Returns: "[\"edge\",\"edge\"]"
```

### Object Operations

```sql
-- Get all keys
jq('{"name":"ekuiper","version":"1.0"}', 'keys')
-- Returns: ["name","version"]

-- Get all values
jq('{"name":"ekuiper","version":"1.0"}', '[.[] ]')
-- Returns: ["ekuiper","1.0"]

-- Check if key exists
jq('{"name":"ekuiper"}', 'has("name")')
-- Returns: true

-- Object construction
jq('{"items":[{"a":1},{"a":2}]}', '.items[] | {b: (.a * 2)}')
-- Returns: "[{\"b\":2},{\"b\":4}]"

-- Delete a field
jq('{"name":{"firstname":"Void","lastname":"King"}}', 'del(.name.firstname)')
-- Returns: {"name":{"lastname":"King"}}

-- Add/modify field
jq('{"name":"ekuiper"}', '.version = "2.0"')
-- Returns: {"name":"ekuiper","version":"2.0"}
```

### Arithmetic and String Operations

```sql
-- Numeric operations
jq('{"num":3}', '.num * 3')
-- Returns: 9

jq('{"a":10,"b":3}', '.a + .b')
-- Returns: 13

jq('{"a":10,"b":3}', '.a - .b')
-- Returns: 7

-- String concatenation
jq('{"str":"hello"}', '.str + " world"')
-- Returns: "hello world"

-- Find indices of substring
jq('"abcb"', 'indices("b")')
-- Returns: [1, 3]

-- String length
jq('"hello"', 'length')
-- Returns: 5
```

### Conditional Logic

```sql
-- If-then-else
jq('{"age":25}', 'if .age >= 18 then "adult" else "minor" end')
-- Returns: "adult"

-- Comparison
jq('{"a":5,"b":3}', '.a > .b')
-- Returns: true

-- Logical operations
jq('{"a":5,"b":3}', '.a > 3 and .b < 5')
-- Returns: true
```

### Edge Cases

```sql
-- Empty object field access returns nil
jq('{}', '.name')
-- Returns: nil

-- Empty array index returns nil
jq('[]', '.[0]')
-- Returns: nil

-- Empty string input returns nil
jq('', '.a')
-- Returns: nil
```
