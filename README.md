## jpoet

Jpoet provides a set of tools that makes it easier to write and reuse Jsonnet code.

### Synopsis

Jsonnet is a powerful and flexible configuration language that extends JSON with advanced programming features.
It supports conditionals, loops, functions, and object-oriented constructs, enabling more concise and reusable configuration code.

In addition to these language features, Jsonnet provides additional tooling.
For example to import external Jsonnet files and to write output to multiple files.
Overall, this promotes modular design and facilitates reuse across different projects, whether authored internally or sourced externally.

However, the standard Jsonnet toolchain does have limitations.
Extending the standard library with custom native functions, for instance, is non-trivial.
It requires developers to create a dedicated Go binary, which must then replace the default Jsonnet CLI to execute configurations.

While the inclusion of additional native functions introduces potential security and side-effect risks, the absence of commonly needed features (such as regular expression support) makes this functionality highly desirable.
A plugin mechanism that allows selective inclusion of safe, useful functions, without having to write new binaries, would significantly improve developer experience.

This is where Jpoet comes into play.
Jpoet introduces a plugin management system built on [go-plugin](https://github.com/hashicorp/go-plugin), the same robust framework used in projects like Terraform and Vault.
With Jpoet, developers can install Jsonnet plugins locally and evaluate configurations via the Jpoet binary.

For detailed usage instructions, refer to the documentation of the respective commands.

### Options

```
  -h, --help   help for jpoet
```

### SEE ALSO

* [jpoet install](#jpoet-install)	 - Jsonnet test is a simple tool to install tests for Jsonnet files
* [jpoet pkg](#jpoet-pkg)	 - Subcommands for building packages and managing them in the target repository
* [jpoet repo](#jpoet-repo)	 - Subcommands for managing target repositories
* [jpoet run](#jpoet-run)	 - Jsonnext run is a simple tool to run Jsonnet files
* [jpoet test](#jpoet-test)	 - Jsonnet test is a simple tool to run tests for Jsonnet files

## jpoet install

Jsonnet test is a simple tool to install tests for Jsonnet files

```
jpoet install [flags] directory
```

### Options

```
  -h, --help   help for install
```

### SEE ALSO

* [jpoet](#jpoet)	 - Jpoet provides a set of tools that makes it easier to write and reuse Jsonnet code.

## jpoet pkg

Subcommands for building packages and managing them in the target repository

### Options

```
  -h, --help   help for pkg
```

### SEE ALSO

* [jpoet](#jpoet)	 - Jpoet provides a set of tools that makes it easier to write and reuse Jsonnet code.
* [jpoet pkg build](#jpoet-pkg-build)	 - Builds Jsonnet packages
* [jpoet pkg push](#jpoet-pkg-push)	 - Pushes Jsonnet packages to the target repository
* [jpoet pkg remove](#jpoet-pkg-remove)	 - Removes Jsonnet packages from the target repository

## jpoet pkg build

Builds Jsonnet packages

```
jpoet pkg build [flags] directory
```

### Options

```
  -b, --build string   The path to the build directory, relative to the package directory (default "build")
  -h, --help           help for build
```

### SEE ALSO

* [jpoet pkg](#jpoet-pkg)	 - Subcommands for building packages and managing them in the target repository

## jpoet pkg push

Pushes Jsonnet packages to the target repository

```
jpoet pkg push [flags] directory
```

### Options

```
  -b, --build string   The path to the build directory, relative to the package directory (default "build")
  -h, --help           help for push
```

### SEE ALSO

* [jpoet pkg](#jpoet-pkg)	 - Subcommands for building packages and managing them in the target repository

## jpoet pkg remove

Removes Jsonnet packages from the target repository

```
jpoet pkg remove [flags] directory
```

### Options

```
  -h, --help   help for remove
```

### SEE ALSO

* [jpoet pkg](#jpoet-pkg)	 - Subcommands for building packages and managing them in the target repository

## jpoet repo

Subcommands for managing target repositories

### Options

```
  -h, --help   help for repo
```

### SEE ALSO

* [jpoet](#jpoet)	 - Jpoet provides a set of tools that makes it easier to write and reuse Jsonnet code.
* [jpoet repo index](#jpoet-repo-index)	 - Indexes Jsonnet repository and updates index README

## jpoet repo index

Indexes Jsonnet repository and updates index README

```
jpoet repo index [flags] directory
```

### Options

```
  -h, --help   help for index
```

### SEE ALSO

* [jpoet repo](#jpoet-repo)	 - Subcommands for managing target repositories

## jpoet run

Jsonnext run is a simple tool to run Jsonnet files

```
jpoet run [flags] filename
```

### Options

```
  -h, --help   help for run
```

### SEE ALSO

* [jpoet](#jpoet)	 - Jpoet provides a set of tools that makes it easier to write and reuse Jsonnet code.

## jpoet test

Jsonnet test is a simple tool to run tests for Jsonnet files

```
jpoet test [flags] directory
```

### Options

```
  -h, --help   help for test
  -j, --json   Outputs the test results in JSON
```

### SEE ALSO

* [jpoet](#jpoet)	 - Jpoet provides a set of tools that makes it easier to write and reuse Jsonnet code.

