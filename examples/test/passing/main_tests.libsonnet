local fibonacci = import './main.libsonnet';

local fibonacciTests = {
  name: 'fibonacci',
  tests: [
    {
      name: '0',
      input:: 0,
      expected: 0,
    },
    {
      name: '1',
      input:: 1,
      expected: 1,
    },
    {
      name: '2',
      input:: 2,
      expected: 1,
    },
    {
      name: '3',
      input:: 3,
      expected: 2,
    },
    {
      name: '4',
      input:: 4,
      expected: 3,
    },
    {
      name: '5',
      input:: 5,
      expected: 5,
    },
    {
      name: '10',
      input:: 10,
      expected: 55,
    },
  ],
};

{
  output(input):: fibonacci(input),
  tests: [
    fibonacciTests,
  ],
}
