local p = import 'pkg/main.libsonnet';

p.pkg({
  source: 'https://github.com/marcbran/jpoet/tree/main/examples/install/plugin',
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'plugin',
  path: 'plugin',
  target: 'p',
  plugins: [
    p.plugin.github('marcbran/jsonnet-plugin-markdown', 'v0.1.0'),
  ],
}, |||
  Test project, please ignore.

  This is a project to test the build and push pipeline.
  The project itself is empty.
|||)
