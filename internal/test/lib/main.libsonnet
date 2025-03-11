local flattenTests(value) =
  if std.objectHasAll(value, 'input') then
    [value]
  else if std.objectHas(value, 'tests') then
    [
      test {
        name: '%s/%s' % [std.get(value, 'name', ''), std.get(test, 'name', '')],
        output::
          if std.get(test, 'output', null) != null then std.get(test, 'output', null)
          else if std.get(value, 'output', null) != null then std.get(value, 'output', null)
          else null,
      }
      for test in std.flattenArrays([flattenTests(test) for test in value.tests])
    ]
  else [];

local applyTests(tests) =
  local results = [
    test {
      actual: if test.output != null then test.output(test.input) else null,
      equal: std.manifestJson(self.actual) == std.manifestJson(test.expected),
    }
    for test in tests
  ];
  {
    results: results,
    passedCount: std.length(std.filter(function(test) test.equal, results)),
    totalCount: std.length(results),
  };

local runTests(tests) = applyTests(flattenTests(tests));

{
  runTests: runTests,
}
