{
  pkg(pkg, description, children={}): {
    coord: {
      repo: pkg.repo,
      branch: pkg.branch,
      path: pkg.path,
    },
    usage: {
      target: pkg.target,
      name: pkg.path,
    },
    source: std.get(pkg, 'source', null),
    description: description,
    children: children,
  },
  desc(description, children={}): {
    description: description,
    children: children,
  },
  ex(examples, children={}): {
    examples: if std.type(examples) == 'array' then examples else [],
    example: if std.type(examples) == 'object' then examples else {},
    children: children,
  },
}
