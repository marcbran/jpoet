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
  test2: p.ex([{
    name: 'Without parameters',
    inputs: [],
    expected: {
      foo: 'bar',
    },
  }, {
    name: 'Markdown format with gensonnet',
    example:
      local g = import 'gensonnet/main.libsonnet';
      g.parseMarkdown('# %s' % [ex.test1.foo]),
  }]),
})
