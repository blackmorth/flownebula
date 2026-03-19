PHP_ARG_ENABLE(flow_nebula, whether to enable flow_nebula support,
[  --enable-flow_nebula           Enable flow_nebula support])

if test "$PHP_FLOW_NEBULA" != "no"; then
  PHP_ADD_INCLUDE($srcdir/probe)
  PHP_NEW_EXTENSION(flow_nebula, probe/nebula_probe.c probe/hooks.c probe/utils.c, $ext_shared)
fi
