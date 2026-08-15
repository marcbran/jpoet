local p = import 'pkg/main.libsonnet';

p.pkg({
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'externalimport',
  path: 'externalimport',
  target: 'ei',
  external: ['ext'],
}, |||
  An externalimport library.

  This will leave imports listed in `external` unresolved, while still
  inlining regular local imports.
|||, {
  test1: p.desc(
    |||
      Test property, sourced from an import listed in `external`.
    |||,
  ),
  test2: p.desc(
    |||
      Another test property, inlined from a local file.
    |||,
  ),
})
