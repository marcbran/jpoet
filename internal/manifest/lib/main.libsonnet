local flattenObject(value, separator='/') =
  if std.type(value) == 'object' then
    std.foldl(function(acc, curr) acc + curr, [
      {
        [std.join(separator, std.filter(function(key) key != '', [child.key, childChild.key]))]: childChild.value
        for childChild in std.objectKeysValues(flattenObject(child.value))
      }
      for child in std.objectKeysValues(value)
    ], {})
  else { '': value };

local runManifest(manifest) =
  flattenObject({
    [kv.key]:
      if std.length(std.findSubstr('.', kv.key)) > 0
      then
        local manifestation = std.get(manifest.manifestations, '.%s' % std.split(kv.key, '.')[1], function(value) value);
        manifestation(kv.value)
      else runManifest(manifest { directory: kv.value })
    for kv in std.objectKeysValues(manifest.directory)
  });

runManifest
