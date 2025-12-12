# pb33f jsonpath

[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](https://pkg.go.dev/github.com/pb33f/jsonpath?tab=doc)

A full implementation of [RFC 9535 JSONPath](https://datatracker.ietf.org/doc/rfc9535/) with **JSONPath Plus** extensions for enhanced querying capabilities.

This library was forked from [speakeasy-api/jsonpath](https://github.com/speakeasy-api/jsonpath).

## What is JSONPath Plus?

JSONPath Plus extends the standard JSONPath specification with powerful context-aware operators, type selectors, and navigation features. These extensions are inspired by and compatible with [JSONPath-Plus/JSONPath](https://github.com/JSONPath-Plus/JSONPath) (the JavaScript reference implementation).

**Key benefits:**
- **100% backward compatible** with RFC 9535 - all standard queries work unchanged
- **Context variables** (`@property`, `@path`, `@parent`, etc.) for advanced filtering
- **Type selectors** (`isString()`, `isNumber()`, etc.) for type-based filtering
- **Parent navigation** (`^`) for traversing up the document tree

## Installation

```bash
go get github.com/pb33f/jsonpath
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/pb33f/jsonpath/pkg/jsonpath"
    "go.yaml.in/yaml/v4"
)

func main() {
    data := `
store:
  book:
    - title: "Book 1"
      price: 10
    - title: "Book 2"
      price: 20
`
    var node yaml.Node
    yaml.Unmarshal([]byte(data), &node)

    // Standard RFC 9535 query
    path, _ := jsonpath.NewPath(`$.store.book[?(@.price > 15)]`)
    results := path.Query(&node)

    // JSONPath Plus query with @property
    path2, _ := jsonpath.NewPath(`$.store.*[?(@property == 'book')]`)
    results2 := path2.Query(&node)
}
```

---

## JSONPath Plus Extensions

### Context Variables

Context variables provide information about the current evaluation context within filter expressions. They are prefixed with `@` and can be used in comparisons.

#### `@property`

Returns the property name (for objects) or index as string (for arrays) used to reach the current node.

```yaml
# Data
paths:
  /users:
    get: { summary: "Get users" }
    post: { summary: "Create user" }
  /orders:
    get: { summary: "Get orders" }
```

```
# Query: Find all GET operations
$.paths.*[?(@property == 'get')]

# Returns: The get objects under /users and /orders
```

#### `@path`

Returns the normalized JSONPath string to the current node being evaluated.

```yaml
# Data
store:
  book:
    - title: "Book 1"
    - title: "Book 2"
```

```
# Query: Find the first book by its path
$.store.book[?(@path == "$['store']['book'][0]")]

# Returns: The first book object
```

#### `@parent`

Returns the parent node of the current node being evaluated. Requires parent tracking to be enabled (automatic when used).

```yaml
# Data
items:
  - name: "Item 1"
    category: "A"
  - name: "Item 2"
    category: "B"
```

```
# Query: Find items where parent is an array
$.items[?(@parent)]

# Returns: All items (parent is the items array)
```

#### `@parentProperty`

Returns the property name or index used to reach the parent of the current node.

```yaml
# Data
store:
  book:
    details: { price: 10 }
  bicycle:
    details: { price: 20 }
```

```
# Query: Find details where parent was reached via 'book'
$.store.*[?(@parentProperty == 'book')]

# Returns: The details object under book
```

#### `@root`

Provides access to the document root from within filter expressions.

```yaml
# Data
config:
  defaultPrice: 10
items:
  - name: "Item 1"
    price: 10
  - name: "Item 2"
    price: 20
```

```
# Query: Find items matching the default price
$.items[?(@.price == @root.config.defaultPrice)]

# Returns: Item 1
```

#### `@index`

Returns the current array index (-1 if not in an array context).

```yaml
# Data
items:
  - name: "First"
  - name: "Second"
  - name: "Third"
```

```
# Query: Find items at even indices
$.items[?(@index == 0 || @index == 2)]

# Returns: First and Third items
```

---

### Type Selector Functions

Type selectors filter nodes based on their data type. They can be used within filter expressions.

| Function | Matches |
|----------|---------|
| `isNull(@)` | Null values |
| `isBoolean(@)` | Boolean values (`true`/`false`) |
| `isNumber(@)` | Numeric values (integers and floats) |
| `isInteger(@)` | Integer values only |
| `isString(@)` | String values |
| `isArray(@)` | Array/sequence nodes |
| `isObject(@)` | Object/mapping nodes |

#### Examples

```yaml
# Data
mixed:
  - 42
  - "hello"
  - true
  - null
  - [1, 2, 3]
  - { key: "value" }
```

```
# Query: Find all string values
$.mixed[?isString(@)]
# Returns: "hello"

# Query: Find all numeric values
$.mixed[?isNumber(@)]
# Returns: 42

# Query: Find all arrays
$.mixed[?isArray(@)]
# Returns: [1, 2, 3]

# Query: Find all objects
$.mixed[?isObject(@)]
# Returns: { key: "value" }
```

---

### Parent Selector (`^`)

The caret operator (`^`) returns the parent of the matched node. This allows you to navigate up the document tree.

```yaml
# Data
store:
  book:
    - title: "Expensive Book"
      price: 100
    - title: "Cheap Book"
      price: 5
```

```
# Query: Find parents of expensive items (price > 50)
$.store.book[?(@.price > 50)]^

# Returns: The book array (parent of the matching book)
```

**Note:** Using `^` on the root node returns an empty result.

---

### Property Name Selector (`~`)

The tilde operator (`~`) returns the property name (key) instead of the value.

```yaml
# Data
person:
  name: "John"
  age: 30
  city: "NYC"
```

```
# Query: Get all property names
$.person.*~

# Returns: ["name", "age", "city"]
```

---

## Standard RFC 9535 Features

This library fully implements RFC 9535, including:

### Selectors

| Selector | Example | Description |
|----------|---------|-------------|
| Root | `$` | The root node |
| Current | `@` | Current node (in filters) |
| Child | `.property` or `['property']` | Direct child access |
| Recursive | `..property` | Descendant search |
| Wildcard | `.*` or `[*]` | All children |
| Array Index | `[0]`, `[-1]` | Specific index (negative from end) |
| Array Slice | `[0:5]`, `[::2]` | Range with optional step |
| Filter | `[?(@.price < 10)]` | Conditional selection |
| Union | `[0,1,2]` or `['a','b']` | Multiple selections |

### Filter Operators

| Operator | Description |
|----------|-------------|
| `==` | Equal |
| `!=` | Not equal |
| `<` | Less than |
| `<=` | Less than or equal |
| `>` | Greater than |
| `>=` | Greater than or equal |
| `&&` | Logical AND |
| `\|\|` | Logical OR |
| `!` | Logical NOT |

### Built-in Functions

| Function | Description |
|----------|-------------|
| `length(@)` | Length of string, array, or object |
| `count(@)` | Number of nodes in a nodelist |
| `match(@.name, 'pattern')` | Regex full match |
| `search(@.name, 'pattern')` | Regex partial match |
| `value(@)` | Extract value from single-node result |

---

## Examples

### Filtering by Property Name

```
# Find all HTTP methods in an OpenAPI spec
$.paths.*[?(@property == 'get' || @property == 'post')]
```

### Complex Path Matching

```
# Find nodes at a specific path pattern
$.store.*.items[*][?(@path == "$['store']['electronics']['items'][0]")]
```

### Type-Safe Queries

```
# Find all string properties in a config
$..config.*[?isString(@)]
```

### Parent Navigation

```
# Get containers of items over $100
$..[?(@.price > 100)]^
```

### Combining Features

```
# Find GET operations where parent path contains 'users'
$.paths[?(@property == '/users')].get
```

---

## ABNF Grammar

The complete ABNF grammar for RFC 9535 JSONPath is available in the [RFC 9535 specification](https://datatracker.ietf.org/doc/rfc9535/).

---

## Contributing

We welcome contributions! Please open a GitHub issue or Pull Request for bug fixes or features.

This library is compliant with the [JSONPath Compliance Test Suite](https://github.com/jsonpath-standard/jsonpath-compliance-test-suite).

---

## License

See [LICENSE](LICENSE) for details.
