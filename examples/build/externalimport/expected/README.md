# externalimport

> An externalimport library.

- [Inlined Code](https://github.com/marcbran/jsonnet/blob/externalimport/externalimport/main.libsonnet): Inlined code published for usage in other projects

## Installation

You can install the library into your project using the [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler):

```shell
jb install https://github.com/marcbran/jsonnet/externalimport@externalimport
```

Then you can import it into your file in order to use it:

```jsonnet
local ei = import 'externalimport/main.libsonnet';
```

## Description

This will leave imports listed in `external` unresolved, while still
inlining regular local imports.

## Fields

### test1

Test property, sourced from an import listed in `external`.

```jsonnet
ei.test1
```


### test2

Another test property, inlined from a local file.

```jsonnet
ei.test2
```

