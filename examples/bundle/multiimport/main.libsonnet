local test1 = import 'lib.libsonnet';
local test2 = import 'lib.libsonnet';

local test3 = {
  foo: 'bar',
};

{
  test1: test1,
  test2: test2,
  test3: test3,
}
