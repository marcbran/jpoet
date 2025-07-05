local test1 =
  local baz = 'baz';

  {
    foo: baz,
  };
local test2 =
  local baz = 'baz';

  {
    foo: baz,
  };

local test3 = {
  foo: 'bar',
};

{
  test1: test1,
  test2: test2,
  test3: test3,
}
