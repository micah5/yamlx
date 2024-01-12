# yamlx

## Introduction
`yamlx` is a Go library that introduces dynamic expressions and loops to yaml.

## Installation
```bash
go get github.com/micah5/yamlx
```

## Example
```yaml
environments: &environments
  - dev
  - prod

subdomains: &subdomains [example, myapp]

servers:
  !for idx, name in *environments:
    - name: *name
      ip: 192.168.1.${(idx + 1) * 100}
      host: ${name}.${join(".", subdomains)}.com
      port: 22
```

You can unmarshal into a struct just like the default yaml library (see [example](examples/test.go)).

Marshalling simply returns a struct back to regular yaml.

## Features

### Expressions
Expressions in `yamlx` allow you to embed dynamic content within your YAML files using the syntax `${<expression>}`

You can use any anchors defined in your code within the brackets (but without the alias prefix, i.e. `*anchor == ${anchor}`)

Other than that, it uses [govaluate](https://github.com/Knetic/govaluate) internally so you can pretty much do anything you can do on there.

**Examples:**
```yaml
dynamicSum: ${ 2 + 3 } # 5
dynamicComparison: ${ 10 > 5 } # true
dynamicConcatenation: ${ "Hello, " + "World!" } # "Hello, World!"
conditionalExpression: ${ 1 == 1 ? "Yes" : "No" } # "Yes"
anchor: &anchor World
anchorExpression: ${ "Hello, " + anchor } # "Hello, World"
```

### Loops
Loops in `yamlx` enable iterating over elements and generating repetitive structures easily. 

You can also use `idx` in your loops to get the index of the current item.

**Examples:**
```yaml
sequence:
  !for number in [1..3]:
    - value: ${number}
  # [{value: 1}, {value: 2}, {value: 3}]

indexedSequence:
  !for idx, number in [10..12]:
    - index: ${idx}
      value: ${number}
  # [{index: 0, value: 10}, {index: 1, value: 11}, {index: 2, value: 12}]
```

### Functions
There are a few functions you can use within expressions. I'll probably add more in the future as a need comes up for them.

#### len
- **API**: `len(string | []any)`
- **Description**: Returns the length of a string or the number of items in a slice.
- **Example**:
  ```yaml
  lengthOfString: ${len("hello")} # 5
  lengthOfArray: ${len([1, 2, 3])} # 3
  ```

#### contains
- **API**: `contains(string, string | []any, any)`
- **Description**: Checks if a string contains a substring or if a slice contains an element.
- **Example**:
  ```yaml
  stringContains: ${contains("hello world", "world")} # true
  arrayContains: ${contains([1, 2, 3], 2)} # true
  ```

#### rand
- **API**: `rand(float64, float64 | []any)`
- **Description**: Generates a random number between two floats or selects a random element from a slice.
- **Example**:
  ```yaml
  randomBetween: ${rand(1, 10)} # (random number between 1 and 10)
  randomElement: ${rand(["apple", "banana", "cherry"])} # (random item from list)
  ```

#### max & min
- **API**: `max(...float64)`, `min(...float64)`
- **Description**: Finds the maximum or minimum value among provided float arguments.
- **Example**:
  ```yaml
  maxValue: ${max(1, 2, 3)} # 3
  minValue: ${min(1, 2, 3)} # 1
  ```

#### upper, lower, title
- **API**: `upper(string)`, `lower(string)`, `title(string)`
- **Description**: Converts a string to uppercase, lowercase, or title case.
- **Example**:
  ```yaml
  upperCase: ${upper("hello")} # "HELLO"
  lowerCase: ${lower("HELLO")} # "hello"
  titleCase: ${title("hello world")} # "Hello World"
  ```

#### trim
- **API**: `trim(string)`
- **Description**: Trims leading and trailing spaces from a string.
- **Example**:
  ```yaml
  trimmedString: ${trim("  hello  ")} # "hello"
  ```

#### join
- **API**: `join(string, []any)`
- **Description**: Joins elements of a slice into a string separated by a specified delimiter.
- **Example**:
  ```yaml
  joinedString: ${join("-", ["a", "b", "c"])} # "a-b-c"
  ```

#### replace
- **API**: `replace(string, string, string)`
- **Description**: Replaces occurrences of a substring within a string.
- **Example**:
  ```yaml
  replacedString: ${replace("hello world", "world", "universe")} # "hello universe"
  ```

#### substr
- **API**: `substr(string, float64, float64)`
- **Description**: Extracts a substring from a given string.
- **Example**:
  ```yaml
  substring: ${substr("hello world", 0, 5)} # "hello"
  ```

#### strrev
- **API**: `strrev(string)`
- **Description**: Reverses a string.
- **Example**:
  ```yaml
  reversedString: ${strrev("hello")} # "olleh"
  ```

#### startswith & endswith
- **API**: `startswith(string, string)`, `endswith(string, string)`
- **Description**: Checks if a string starts or ends with a specified substring.
- **Example**:
  ```yaml
  startsWith: ${startswith("hello world", "hello")} # true
  endsWith: ${endswith("hello world", "world")} # true
  ```

#### alltrue & anytrue
- **API**: `alltrue(...bool)`, `anytrue(...bool)`
- **Description**: Evaluates if all or any of the provided boolean values are true.
- **Example**:
  ```yaml
  allAreTrue: ${alltrue(true, true, false)} # false
  anyAreTrue: ${anytrue(false, false, true)} # true
  ```
