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
By default, context tracking is eager to preserve historical behavior. If you want to reduce overhead for queries that do not use context variables, enable `config.WithLazyContextTracking()` to turn on tracking only when a query uses `@property`, `@path`, `@parentProperty`, or `@index`.

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

## Spectral Compatibility

Spectral rulesets use a safe JavaScript-flavored JSONPath subset that is not
part of RFC 9535. Enable it explicitly when executing Spectral `given`
selectors:

```go
import (
    "github.com/pb33f/jsonpath/pkg/jsonpath"
    "github.com/pb33f/jsonpath/pkg/jsonpath/config"
)

path, err := jsonpath.NewPath(
    `$.paths[?(@property && !@property.match(/\/openapi\.json/))]~`,
    config.WithSpectralCompatibility(),
)
```

`WithSpectralCompatibility()` implies the property-name extension and all
JSONPath Plus features, including context variables, `~`, and `^`.
`WithPropertyNameExtension()` may still be supplied and is redundant but
valid. `WithLazyContextTracking()` is independent and produces the same
results in eager and lazy modes.

Spectral compatibility and `WithStrictRFC9535()` are mutually exclusive.
Supplying both returns a configuration error regardless of option order. The
default dialect remains unchanged.

### Compatibility matrix

| Status | Feature | Behavior |
|--------|---------|----------|
| Supported | Truthiness | Missing values, `null`, `false`, zero, and empty strings are false; arrays, objects, non-empty strings, and non-zero numbers are true |
| Supported | Regex literals | `/pattern/flags`, compiled once with Go's linear-time RE2 engine |
| Supported | `value.match(regex)` | Partial regex search on strings; usable only as a truthy test |
| Supported | `value.indexOf(string[, fromIndex])` | JavaScript UTF-16 code-unit indexing on strings |
| Supported | `value.includes(value[, fromIndex])` | Safe membership checks on strings and sequences |
| Supported | `value.length` | UTF-16 length for strings and element count for sequences |
| Supported | `value.constructor.name` | Read-only synthetic type metadata: `Array`, `Object`, `String`, `Number`, or `Boolean` |
| Supported | `void 0` | The missing/undefined value used by official Spectral selectors |
| Supported | `===` and `!==` | JavaScript strict equality, including reference identity for arrays and objects |
| Supported | `==` and `!=` | JavaScript scalar coercion and reference identity for arrays and objects |
| Partial by design | `value.match(regex)` result | Truthiness is exact; comparison and result chaining are rejected because the evaluator does not expose JavaScript match arrays |
| Partial by design | Regex `u` flag | Accepted for RE2-compatible patterns; JavaScript-only code-point escapes are rejected |
| Unsupported | Arbitrary JavaScript and methods | Assignments, statements, functions, unknown methods, constructors, prototype access, and reflection are rejected |
| Unsupported | JavaScript-only regex features | Lookaround, backreferences, `d`, `v`, and `y` flags are rejected |

RFC functions remain available in Spectral mode. For example,
`match(@.name, 'a.*')` is RFC 9535 full-string matching, whereas
`@.name.match(/a/)` is a Spectral partial search. Spectral `match` returns an
internal boolean-compatible result; comparing it, reading its length, or
chaining from it is rejected rather than approximated.

### Context value types

Spectral mode follows JavaScript object and array property behavior:

- `@property` and `@parentProperty` are strings for mapping entries. YAML keys
  are stringified, so an unquoted response key such as `404` works with
  `@property.match(/^(4|5)/)`.
- They are numbers for sequence-index comparisons. `@property === 3` can match
  the fourth element, while `@property === '3'` cannot. Nimma nevertheless
  stringifies a sequence property for its whitelisted `match`, `indexOf`, and
  `includes` operations; this engine-specific behavior is pinned in the
  differential corpus.
- A method or synthetic property used on the wrong value type evaluates as a
  non-match and never panics.

### Regex dialect and safety

Regex flags `i`, `m`, and `s` map to RE2 modes. `g` is accepted because it does
not change boolean match results. `u` is accepted only when the pattern does
not require JavaScript-only Unicode syntax such as `\u{...}`. Flags `d`, `v`,
and `y`, duplicate or unknown flags, lookaround, and backreferences are
rejected during compilation with line and column information.

Spectral mode does not execute JavaScript and does not use reflection. It
rejects arbitrary methods, assignments, statements, constructors,
`__proto__`, prototype traversal, computed constructor access, getters, and
setters before document traversal. Only the operations listed above are
available.

### Compatibility baseline

The checked corpus in
[`pkg/jsonpath/testdata/spectral`](pkg/jsonpath/testdata/spectral) records the
pinned Spectral version, source selector, exact normalized result paths,
multiplicity, and whether Spectral selected Nimma or its JSONPath Plus
fallback. It includes every selector collected from the pinned official
OpenAPI and AsyncAPI rulesets plus focused migration, public-ruleset, Unicode,
security, and Vacuum issue fixtures. A separate public inventory classifies 51
deduplicated selectors from four pinned rulesets. Public Spectral aliases are
expanded before classification, and every source records its repository,
commit, path, and license identifier.

Regenerate the differential results and official-selector inventory with:

```bash
cd pkg/jsonpath/testdata/spectral/generate
npm ci
npm run collect-official
npm run collect-public
npm run generate
```

The currently supported boundary is documented above. Arbitrary JavaScript,
user-provided functions, prototype access, and unsupported regex constructs
are intentionally unsupported.

---

## Overlay Support

This library includes support for YAML overlays, which allow you to apply structured modifications to YAML documents using JSONPath expressions.

### Basic Overlay Usage

Overlays are defined in YAML format and specify actions to apply to a target document:

```yaml
overlay: "1.0.0"
info:
  title: My Overlay
  version: "1.0"
actions:
  - target: $.spec.replicas
    update: 3
```

```go
package main

import (
    "fmt"
    "github.com/pb33f/jsonpath/pkg/overlay"
    "go.yaml.in/yaml/v4"
)

func main() {
    // Parse the overlay
    ov, _ := overlay.ParseOverlay(overlayYAML)
    
    // Apply to a document
    result, _ := ov.Apply(documentNode)
}
```

### Upsert Action

The `upsert` action combines update and insert behavior. When `upsert: true` is set:

- If the target path exists, the value is updated (same as `update`)
- If the target path doesn't exist, the path is created and the value is set

**Example: Create nested paths**

```yaml
# Input document
spec:
  existing: value

# Overlay
overlay: "1.0.0"
actions:
  - target: $.spec.config.nested.key
    update: created
    upsert: true

# Result
spec:
  existing: value
  config:
    nested:
      key: created
```

**Example: Update existing values**

```yaml
# Input document
spec:
  existing: value

# Overlay
overlay: "1.0.0"
actions:
  - target: $.spec.existing
    update: updated
    upsert: true

# Result
spec:
  existing: updated
```

**Example: Array elements**

```yaml
# Input document
items:
  - name: first

# Overlay
overlay: "1.0.0"
actions:
  - target: $.items[0].name
    update: updated
    upsert: true

# Result
items:
  - name: updated
```

### Supported Path Types for Upsert

Upsert works with **singular paths** - paths that resolve to exactly one location:

| Path Type | Example | Behavior |
|-----------|---------|----------|
| Member name | `$.foo.bar` | Creates nested maps as needed |
| Array index | `$.items[0]` | Creates arrays and sets at index |
| Combined | `$.a.b[2].c` | Creates nested structures |

### Unsupported Path Types

The following path types **cannot** be used with upsert (will return an error):

| Path Type | Example | Reason |
|-----------|---------|--------|
| Wildcard | `$.*.foo` | Ambiguous - which child? |
| Recursive descent | `$..foo` | Ambiguous location |
| Filter | `$[?(@.x)]` | Query, not specific location |
| Multiple selectors | `$['a','b']` | Multiple locations |
| Slice | `$[0:5]` | Multiple locations |

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
