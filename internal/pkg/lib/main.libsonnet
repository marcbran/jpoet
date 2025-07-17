local g = import 'gensonnet/main.libsonnet';
local md = import 'markdown/main.libsonnet';

local stripLocRange(obj) =
  if std.type(obj) == 'object' then
    { [kv.key]: stripLocRange(kv.value) for kv in std.objectKeysValues(obj) if kv.key != 'locRange' }
  else if std.type(obj) == 'array' then
    [stripLocRange(elem) for elem in obj]
  else obj;

local invoke(func, params) =
  if std.length(params) == 0 then func()
  else if std.length(params) == 1 then func(params[0])
  else if std.length(params) == 2 then func(params[0], params[1])
  else if std.length(params) == 3 then func(params[0], params[1], params[2])
  else if std.length(params) == 4 then func(params[0], params[1], params[2], params[3])
  else if std.length(params) == 5 then func(params[0], params[1], params[2], params[3], params[4])
  else if std.length(params) == 6 then func(params[0], params[1], params[2], params[3], params[4], params[5]);

local getIndex(array, index, default=null) =
  if std.length(array) > index then array[index] else default;

local getDeep(value, indices, default=null) =
  if std.length(indices) == 0 then value
  else
    local index = indices[0];
    if std.type(value) == 'object' && std.objectHas(value, index) then getDeep(value[index], indices[1:], default)
    else if std.type(value) == 'array' && std.type(index) == 'number' && std.length(value) > index then getDeep(value[index], indices[1:], default)
    else if std.type(value) == 'array' && std.type(index) == 'function' then getDeep([val for val in value if index(val)], indices[1:], default)
    else default;

local injectExampleString(examples, examplesNode) =
  if examples == null then null
  else if std.type(examplesNode) == 'string' then injectExampleString(examples, g.parseJsonnet(examplesNode))
  else if examplesNode.__kind__ == 'Local' then injectExampleString(examples, examplesNode.body)
  else if examplesNode.__kind__ == 'Apply' then

    local injectSingleExampleString(example, exampleNode) =
      local node = getDeep(exampleNode, ['expr', 'fields', function(field) field.id == 'example', 0, 'expr2'], null);
      example + if node != null then { string: g.manifestJsonnet(node) } else {};

    local injectArrayExampleString(examples, exampleNodes) =
      std.mapWithIndex(
        function(index, example) injectSingleExampleString(example, exampleNodes[index]),
        examples
      );

    examples {
      example: injectSingleExampleString(examples.example, getDeep(examplesNode.arguments.positional, [0], {})),
      examples: injectArrayExampleString(examples.examples, getDeep(examplesNode.arguments.positional, [0, 'expr', 'elements'], [])),
      ex: {
        children: {
          [field.id]: injectExampleString(examples.ex.children[field.id], field.expr2)
          for field in getDeep(examplesNode.arguments.positional, [1, 'expr', 'fields'], [])
        },
      },
    }
  else examplesNode;

local merge(lib, desc, examples, coord, usage, source) = {
  type: std.type(lib),
  implementation:: lib,
  coord: coord,
  usage: usage,
  source: source,
  description: std.get(desc, 'description', ''),
  examples: if std.type(examples) == 'object' then std.get(examples, 'examples', []) else [],
  example: if std.type(examples) == 'object' then std.get(examples, 'example', {}) else {},
  children: [
    merge(
      std.get(lib, key, null),
      getDeep(desc, ['children', key], null),
      getDeep(examples, ['children', key], null),
      coord,
      {
        target: '%s.%s' % [usage.target, key],
        name: key,
      },
      source
    )
    for key in std.objectFields(desc.children)
  ],
};

local mergeRoot(lib, pkg, examples) =
  merge(lib, pkg, examples, pkg.coord, pkg.usage, pkg.source) + { root: true };

local heading(elem, depth) = [
  md.Heading(depth, elem.usage.name),
];

local summary(elem) = [
  md.Paragraph(std.split(elem.description, '\n')[0]),
];

local description(elem) = [
  md.Paragraph(std.split(elem.description, '\n')[2:]),
];

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
    ||| % [elem.coord.repo, elem.coord.path, elem.coord.branch], language='shell'
  ),
  md.Paragraph('Then you can import it into your file in order to use it:'),
  md.FencedCodeBlock(
    |||
      local %s = import '%s/main.libsonnet';
    ||| % [elem.usage.target, elem.usage.name], language='jsonnet'
  ),
];

local summarySection(elem) =
  local httpRepo =
    if std.endsWith(elem.coord.repo, '.git')
    then elem.coord.repo[:std.length(elem.coord.repo) - 4]
    else elem.coord.repo;
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
          md.Link('Inlined Code', '%s/blob/%s/%s/main.libsonnet' % [httpRepo, elem.coord.branch, elem.coord.path]),
          ': Inlined code published for usage in other projects',
        ]),
      ]),
    ]),
  ];

local descriptionSection(elem, depth) =
  local lines = std.split(elem.description, '\n');
  if std.length(lines) > 0 then [
    md.Heading(depth, 'Description'),
    md.Paragraph(std.split(elem.description, '\n')[2:]),
  ] else [];

local usage(elem) = [
  md.FencedCodeBlock(
    if elem.type == 'function' then
      '%s()' % elem.usage.target
    else
      '%s' % elem.usage.target,
    language='jsonnet'
  ),
];

local example(example, usage, implementation, depth) =
  if std.objectHas(example, 'inputs') || std.objectHas(example, 'string') then
    [if std.objectHas(example, 'name') then md.Heading(depth, example.name) else ''] +
    (if std.objectHas(example, 'inputs') then [
       md.Paragraph([md.Strong('Calling')]),
       md.FencedCodeBlock(
         |||
           %s(%s)
         ||| % [usage.target, std.join(', ', [std.manifestJson(input) for input in example.inputs])],
         language='jsonnet'
       ),
     ] else [
       md.Paragraph([md.Strong('Running')]),
       md.FencedCodeBlock(
         |||
           local %s = import '%s/main.libsonnet';
           &nbsp;
         ||| % [usage.target, usage.name] +
         example.string,
         language='jsonnet'
       ),
     ]) + [
      md.Paragraph([md.Strong('yields')]),
      md.FencedCodeBlock(
        if std.objectHas(example, 'output') then
          example.output
        else
          std.manifestJson(
            if std.objectHas(example, 'inputs') then
              invoke(implementation, example.inputs)
            else
              example.example
          ),
        language='json'
      ),
    ]
  else [];

local exampleList(examples, usage, implementation, depth) =
  if std.length(examples) > 0 then
    [md.Heading(depth, 'Examples')] +
    [
      example(ex, usage, implementation, depth + 1)
      for ex in examples
    ]
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
    + example(elem.example, elem.usage, elem.implementation, depth + 1)
    + install(elem, depth + 1)
    + descriptionSection(elem, depth + 1)
    + fields(elem, depth + 1)
  else
    heading(elem, depth)
    + summary(elem)
    + usage(elem)
    + description(elem)
    + example(elem.example { name: 'Example' }, elem.usage, elem.implementation, depth + 1)
    + exampleList(elem.examples, elem.usage, elem.implementation, depth + 1)
    + std.flattenArrays([documentation(child, depth + 1) for child in elem.children]);

local manifest(lib, pkg, examples, examplesString) =
  local elem = mergeRoot(lib, pkg, injectExampleString(examples, examplesString));
  local doc = documentation(elem);
  {
    directory: {
      'README.md': md.Document(doc),
    },
  };

manifest
