local romanNum = import './main.libsonnet';

local romanNumTests = {
  name: 'romanNum',
  tests: [
    {
      name: '1',
      input:: 1,
      expected: 'I',
    },
    {
      name: '2',
      input:: 2,
      expected: 'II',
    },
    {
      name: '3',
      input:: 3,
      expected: 'III',
    },
    {
      name: '4',
      input:: 4,
      expected: 'IV',
    },
    {
      name: '5',
      input:: 5,
      expected: 'V',
    },
  ],
};

{
  output(input):: romanNum(input),
  tests: [
    romanNumTests,
  ],
}
