local g = import 'gensonnet/main.libsonnet';
local md = import 'markdown/main.libsonnet';

local pkg(name, contents, depth) =
  local summary = g.parseMarkdown(contents)[1:4];
  [
    [
      summary[0][0],
      summary[0][1] {
        level: depth,
      },
      summary[0][2],
    ],
    summary[1],
    summary[2][0:2] +
    [
      md.ListItem([
        md.Paragraph([
          md.Link('Readme', name),
          ': Documentation of installation and usage',
        ]),
      ]),
    ] + summary[2][2:],
  ];

local index(files, depth=1) = md.Document([
  md.Heading1('Jsonnet'),
  md.Paragraph(|||
    This repository hosts reusable Jsonnet packages prepared for easy integration in other projects.
    Each package has been processed by inlining all its code into a single file.
    Additionally, each output file has been placed onto a separate branch.
    This enables lightweight and straightforward usage with the [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler).
  |||),
] + std.flattenArrays([
  pkg(kv.key, kv.value, 2)
  for kv in std.objectKeysValues(files)
  if std.endsWith(kv.key, '/README.md')
]));

local manifest(files) = {
  directory: {
    'README.md': index(files),
  },
};

manifest
