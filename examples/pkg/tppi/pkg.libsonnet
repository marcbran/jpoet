local p = import 'jsonnet-pkg/main.libsonnet';

p.pkg({
  source: 'https://github.com/marcbran/devsonnet/tree/main/examples/pkg/tppi',
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'tppi',
  path: 'tppi',
  target: 'tppi',
}, |||
  Test project, please ignore.

  This is a project to test the build and push pipeline.
|||)
