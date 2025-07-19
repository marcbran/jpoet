# examples

> A examples library.

- [Inlined Code](https://github.com/marcbran/jsonnet/blob/examples/examples/main.libsonnet): Inlined code published for usage in other projects

## Installation

You can install the library into your project using the [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler):

```shell
jb install https://github.com/marcbran/jsonnet.git/examples@examples
```

Then you can import it into your file in order to use it:

```jsonnet
local ex = import 'examples/main.libsonnet';
```

## Description

This is to show off how examples are included in the README

## Fields

### test1

Test property.

```jsonnet
ex.test1
```


#### Example

**Running**

```jsonnet
local ex = import 'test1/main.libsonnet';
&nbsp;
ex.test1
```

**yields**

```json
{
    "foo": "bar"
}
```

### test2

Test function property.

```jsonnet
ex.test2()
```


#### Examples

##### Without parameters

**Calling**

```jsonnet
ex.test2()
```

**yields**

```json
{
    "foo": "bar"
}
```

##### Markdown format with gensonnet

**Running**

```jsonnet
local ex = import 'test2/main.libsonnet';
&nbsp;
local g = import 'gensonnet/main.libsonnet';
g.parseMarkdown('# %s' % [ex.test1.foo])
```

**yields**

```json
[
    "Document",
    [
        "Heading",
        {
            "level": 1
        },
        "bar"
    ]
]
```
