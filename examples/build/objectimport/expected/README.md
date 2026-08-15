# objectimport

> An objectimport library.

- [Inlined Code](https://github.com/marcbran/jsonnet/blob/objectimport/objectimport/main.libsonnet): Inlined code published for usage in other projects

## Installation

You can install the library into your project using the [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler):

```shell
jb install https://github.com/marcbran/jsonnet/objectimport@objectimport
```

Then you can import it into your file in order to use it:

```jsonnet
local oi = import 'objectimport/main.libsonnet';
```

## Description

This will inline imports that are used as object field values directly,
not just imports assigned through a chain of locals.

## Fields

### test1

Test property.

```jsonnet
oi.test1
```


### test2

Another test property.

```jsonnet
oi.test2
```

