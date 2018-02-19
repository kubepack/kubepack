local shardTemplate = import "shard.jsonnet.TEMPLATE";

shardTemplate + {
  name:: "foocorp",
  namespace:: "default",
}
