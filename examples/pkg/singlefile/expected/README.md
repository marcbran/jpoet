# singlefile

A singlefile library.

This should just copy the library as-is and create a README file for it.

## Installation

You can install the library into your project using the [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler):

```shell
jb install https://github.com/marcbran/jsonnet.git/singlefile@singlefile
```

Then you can import it into your file in order to use it:

```jsonnet
local sf = import 'singlefile/main.libsonnet';
```

## Fields

### test1

Test property.

```jsonnet
sf.test1
```

