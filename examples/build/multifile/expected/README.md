# multifile

> A multifile library.

- [Inlined Code](https://github.com/marcbran/jsonnet/blob/multifile/multifile/main.libsonnet): Inlined code published for usage in other projects

## Installation

You can install the library into your project using the [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler):

```shell
jb install https://github.com/marcbran/jsonnet/multifile@multifile
```

Then you can import it into your file in order to use it:

```jsonnet
local mf = import 'multifile/main.libsonnet';
```

## Description

This will inline all the imports and create the readme.

## Fields

### test1

Test property.

```jsonnet
mf.test1
```


### test2

Another test property.

```jsonnet
mf.test2
```

