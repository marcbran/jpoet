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

local merge(lib, desc, examples, coord, usage) = {
  type: std.type(lib),
  implementation:: lib,
  coord: coord,
  usage: usage,
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
        path: '%s.%s' % [usage.path, key],
        name: key,
      }
    )
    for key in std.objectFields(desc.children)
  ],
};

local mergeRoot(lib, pkg, examples) =
  merge(lib, pkg, examples, pkg.coord, pkg.usage) + { root: true };

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
    ||| % [elem.usage.path, elem.usage.name], language='jsonnet'
  ),
];

local usage(elem) = [
  md.FencedCodeBlock(
    if elem.type == 'function' then
      '%s()' % elem.usage.path
    else
      '%s' % elem.usage.path,
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
         ||| % [usage.path, std.join(', ', [std.manifestJson(input) for input in example.inputs])],
         language='jsonnet'
       ),
     ] else [
       md.Paragraph([md.Strong('Running')]),
       md.FencedCodeBlock(
         |||
           local %s = import '%s/main.libsonnet';
           &nbsp;
         ||| % [usage.path, usage.name] +
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
  if std.get(elem, 'root', false) then
    heading(elem, depth)
    + summary(elem)
    + example(elem.example, elem.usage, elem.implementation, depth + 1)
    + description(elem)
    + install(elem, depth + 1)
    + [md.Heading(depth + 1, 'Fields')]
    + std.flattenArrays([documentation(child, depth + 2) for child in elem.children])
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
