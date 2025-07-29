local p = import 'pkg/main.libsonnet';

p.pkg({
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'multiimport',
  path: 'multiimport',
  target: 'mi',
}, |||
  A multiimport library.

  This will inline all the imports and create the readme.
|||, {
  test1: p.desc(
    |||
      Test property.
    |||,
  ),
  test2: p.desc(
    |||
      Another test property.
    |||,
  ),
  test3: p.desc(
    |||
      One more test property.
    |||,
  ),
})
