{
  pkg(pkg, description, children={}): {
    coordinates: {
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
    plugins: pkg.plugins,
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
  plugin: {
    github(repo, version): {
      github: {
        repo: repo,
        version: version,
      },
    },
  },
}
