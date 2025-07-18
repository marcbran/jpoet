local p = import 'jsonnet-pkg/main.libsonnet';

p.pkg({
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'examples',
  path: 'examples',
  target: 'ex',
}, |||
  A examples library.

  This is to show off how examples are included in the README
|||, {
  test1: p.desc(
    |||
      Test property.
    |||,
  ),
})
