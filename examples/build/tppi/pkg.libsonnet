local p = import 'pkg/main.libsonnet';

p.pkg({
  source: 'https://github.com/marcbran/jpoet/tree/main/examples/build/tppi',
  repo: 'git@github.com:marcbran/jsonnet.git',
  branch: 'tppi',
  path: 'tppi',
  target: 'tppi',
}, |||
  Test project, please ignore.

  This is a project to test the build and push pipeline.
  The project itself is empty.
|||)
