local test1 = import 'lib.libsonnet';

local test2 = {
  foo: 'bar',
};

{
  test1: test1,
  test2: test2,
}
