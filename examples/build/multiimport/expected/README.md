# multiimport

> A multiimport library.

- [Inlined Code](https://github.com/marcbran/jsonnet/blob/multiimport/multiimport/main.libsonnet): Inlined code published for usage in other projects

## Installation

You can install the library into your project using the [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler):

```shell
jb install https://github.com/marcbran/jsonnet/multiimport@multiimport
```

Then you can import it into your file in order to use it:

```jsonnet
local mi = import 'multiimport/main.libsonnet';
```

## Description

This will inline all the imports and create the readme.

## Fields

### test1

Test property.

```jsonnet
mi.test1
```


### test2

Another test property.

```jsonnet
mi.test2
```


### test3

One more test property.

```jsonnet
mi.test3
```

