local p = import 'pkg/main.libsonnet';

p.pkg({
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'singlefile',
  path: 'singlefile',
  target: 'sf',
}, |||
  A singlefile library.

  This should just copy the library as-is and create a README file for it.
|||, {
  test1: p.desc(
    |||
      Test property.
    |||,
  ),
})
