# jpoet

A Swiss Army knife of small, focused tools for working with [Jsonnet](https://jsonnet.org/).
Each one handles a single task, mostly aimed at making CI workflows or local development easier.

Available tools:

- **test**: runs table-driven tests for Jsonnet code
- **build**: builds inlined Jsonnet code and package README
- **push**: pushes Jsonnet code to a separate GitHub repository

## Installation

To install jpoet, first ensure you have Go installed on your system. Then, run:

```shell
go install github.com/marcbran/jpoet@latest
```

This command will fetch and install the latest version of jpoet.

## Tools

Below is an explanation of each tool.

### test

Runs table-driven tests for Jsonnet code.

#### Motivation

Writing more Jsonnet code made it clear that having a way to test things was becoming important.
Since Jsonnet is side-effect free, it's a great match for table-driven tests, which are also common in Go.
Tests like these started showing up across a few different projects, often using similar scripts to run them.
This test tool is just a Go-based version of those scripts — built to make things more stable, consistent, and easier to use.

#### Usage

1. Write your Jsonnet code tests in the expected test-table format.
   You can see examples of this format in the `./examples/test` folder.
2. Put these tests into files of the format `*_tests.libsonnet`.
3. Run the test command.
   ```shell
   # Runs the tests in the current folder
   $ jpoet test
   
   # Runs the tests in the specified folder
   $ jpoet test ./examples/tests
   
   # Runs the tests with JSON output
   $ jpoet test --json
   ```
   
   If all the tests are passing, the test command exits with `0` and just outputs the number of passed tests:
   ```
   Passed: 7/7
   ```

   If some of the tests are failing, the test command exits with `1` and also outputs details about the failing tests:
   ```
       Name: main_tests.libsonnet/romanNum/4
     Actual: IIII
   Expected: IV

       Name: main_tests.libsonnet/romanNum/5
     Actual: IIIII
   Expected: V

     Passed: 3/5
   ```

   When provided the `--json` flag, the command will output everything as machine-readable JSON.
   This is useful in case some of the test values are to be used for other tests such as integration tests.

### build

#### Motivation

`TODO`

#### Usage

`TODO`

### push

#### Motivation

`TODO`

#### Usage

`TODO`

## License

jpoet is licensed under the Apache License 2.0. See the [LICENSE](./LICENSE) file for details.
