local ext = import 'ext';
local test1 = { value: ext };
local test2 = import 'lib.libsonnet';

{
  test1: test1,
  test2: test2,
}
