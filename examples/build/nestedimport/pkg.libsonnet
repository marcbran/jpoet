local p = import 'pkg/main.libsonnet';

p.pkg({
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'nestedimport',
  path: 'nestedimport',
  target: 'ni',
}, |||
  A nestedimport library.

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
})
