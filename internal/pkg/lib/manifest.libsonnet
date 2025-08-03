local resolvePkgConfig = import './resolve_pkg_config.libsonnet';
local j = import 'jsonnet/main.libsonnet';
local md = import 'markdown/main.libsonnet';

local invoke(func, params) =
  if std.length(params) == 0 then func()
  else if std.length(params) == 1 then func(params[0])
  else if std.length(params) == 2 then func(params[0], params[1])
  else if std.length(params) == 3 then func(params[0], params[1], params[2])
  else if std.length(params) == 4 then func(params[0], params[1], params[2], params[3])
  else if std.length(params) == 5 then func(params[0], params[1], params[2], params[3], params[4])
  else if std.length(params) == 6 then func(params[0], params[1], params[2], params[3], params[4], params[5]);

local heading(elem, depth) = [
  md.Heading(depth, elem.usage.name),
];

local summary(elem) = [
  md.Paragraph(std.split(elem.description, '\n')[0]),
];

local description(elem) = [
  md.Paragraph(std.join('\n', std.split(elem.description, '\n')[2:])),
];

local summarySection(elem) =
  local httpRepo =
    if std.endsWith(elem.coordinates.repo, '.git')
    then elem.coordinates.repo[:std.length(elem.coordinates.repo) - 4]
    else elem.coordinates.repo;
  [
    md.Blockquote([md.Paragraph(std.split(elem.description, '\n')[0])]),
    md.List('-', 0, (
      if std.objectHas(elem, 'source') && elem.source != null then [
        md.ListItem([
          md.Paragraph([
            md.Link('Source Code', elem.source),
            ': Original source code',
          ]),
        ]),
      ] else []
    ) + [
      md.ListItem([
        md.Paragraph([
          md.Link('Inlined Code', '%s/blob/%s/%s/main.libsonnet' % [httpRepo, elem.coordinates.branch, elem.coordinates.path]),
          ': Inlined code published for usage in other projects',
        ]),
      ]),
    ]),
  ];

local descriptionSection(elem, depth) =
  local lines = std.split(elem.description, '\n');
  if std.length(lines) > 0 then [
    md.Heading(depth, 'Description'),
    md.Paragraph(std.join('\n', std.split(elem.description, '\n')[2:])),
  ] else [];

local install(elem, depth) = [
  md.Heading(depth, 'Installation'),
  md.Paragraph([
    'You can install the library into your project using the ',
    md.Link('jsonnet-bundler', 'https://github.com/jsonnet-bundler/jsonnet-bundler'),
    ':',
  ]),
  md.FencedCodeBlock(
    |||
      jb install %s/%s@%s
    ||| % [elem.coordinates.repo, elem.coordinates.path, elem.coordinates.branch], language='shell'
  ),
  md.Paragraph('Then you can import it into your file in order to use it:'),
  md.FencedCodeBlock(
    |||
      local %s = import '%s/main.libsonnet';
    ||| % [elem.usage.target, elem.usage.name], language='jsonnet'
  ),
];

local usage(elem) = [
  md.FencedCodeBlock(
    if elem.type == 'function' then
      '%s()' % elem.usage.target
    else
      '%s' % elem.usage.target,
    language='jsonnet'
  ),
];

local example(example, coordinates, usage, implementation, depth) =
  if std.objectHas(example, 'inputs') || std.objectHas(example, 'string') then
    [if std.objectHas(example, 'name') then md.Heading(depth, example.name) else ''] +
    (if std.objectHas(example, 'inputs') then [
       md.Heading(depth + 1, 'Calling'),
       md.FencedCodeBlock(
         |||
           %s(%s)
         ||| % [usage.target, std.join(', ', [std.manifestJson(input) for input in example.inputs])],
         language='jsonnet'
       ),
     ] else [
       md.Heading(depth + 1, 'Running'),
       md.FencedCodeBlock(
         |||
           local %s = import '%s/main.libsonnet';
         ||| % [std.split(usage.target, '.')[0], coordinates.path] +
         example.string,
         language='jsonnet'
       ),
     ]) + [
      md.Heading(depth + 1, 'yields'),
      local output =
        if std.objectHas(example, 'output') then
          example.output
        else
          if std.objectHas(example, 'inputs') then
            invoke(implementation, example.inputs)
          else
            example.example;
      md.FencedCodeBlock(
        if std.type(output) == 'string' then output else std.manifestJson(output),
        if std.type(output) == 'string' then '' else 'json'
      ),
    ]
  else [];

local exampleList(examples, coordinates, usage, implementation, depth) =
  if std.length(examples) > 0 then
    [md.Heading(depth, 'Examples')] +
    std.flattenArrays([
      example(ex, coordinates, usage, implementation, depth + 1)
      for ex in examples
    ])
  else [];

local documentation(elem, depth=1) =
  local fields(elem, depth) =
    if std.length(elem.children) > 0 then
      [md.Heading(depth, 'Fields')]
      + std.flattenArrays([documentation(child, depth + 1) for child in elem.children])
    else [];

  if std.get(elem, 'root', false) then
    heading(elem, depth)
    + summarySection(elem)
    + example(elem.example, elem.coordinates, elem.usage, elem.implementation, depth + 1)
    + install(elem, depth + 1)
    + descriptionSection(elem, depth + 1)
    + fields(elem, depth + 1)
  else
    heading(elem, depth)
    + summary(elem)
    + usage(elem)
    + description(elem)
    + example(elem.example { name: 'Example' }, elem.coordinates, elem.usage, elem.implementation, depth + 1)
    + exampleList(elem.examples, elem.coordinates, elem.usage, elem.implementation, depth + 1)
    + std.flattenArrays([documentation(child, depth + 1) for child in elem.children]);

local manifest(lib, libString, pkg, examples, examplesString) =
  local pkgConfig = resolvePkgConfig(lib, pkg, examples, examplesString);
  local doc = documentation(pkgConfig);
  {
    'main.libsonnet': j.manifestJsonnet(j.parseJsonnet(libString)),
    'README.md': md.manifestMarkdown(md.Document(doc)),
  };

manifest
