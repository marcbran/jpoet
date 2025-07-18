local ex = import 'input/lib.libsonnet';
local p = import 'jsonnet-pkg/main.libsonnet';

p.ex({
}, {
  test1: p.ex({
    example: ex.test1,
    expected: {
      foo: 'bar',
    },
  }),
})
