local p = import 'pkg/main.libsonnet';

p.pkg({
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'objectimport',
  path: 'objectimport',
  target: 'oi',
}, |||
  An objectimport library.

  This will inline imports that are used as object field values directly,
  not just imports assigned through a chain of locals.
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
