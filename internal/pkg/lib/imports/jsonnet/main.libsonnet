{
  manifestJsonnet(jsonnet): std.native('invoke:jsonnet')('manifestJsonnet', [jsonnet]),
  parseJsonnet(jsonnet): std.native('invoke:jsonnet')('parseJsonnet', [jsonnet]),
}
