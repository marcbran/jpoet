local test1 = import 'lib2.libsonnet';

local test2 = {
  foo: 'bar',
};

{
  test1: test1,
  test2: test2,
}
