local ex = import './main.libsonnet';
local p = import 'pkg/main.libsonnet';

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
    name: 'Markdown format with plugin',
    example:
      local md = import 'markdown/main.libsonnet';
      md.parseMarkdown('# %s' % [ex.test1.foo]),
  }]),
})
