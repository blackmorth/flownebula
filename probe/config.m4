PHP_ARG_ENABLE(nebula_probe, whether to enable nebula_probe support,
[  --enable-nebula_probe           Enable nebula_probe support])

if test "$PHP_NEBULA_PROBE" != "no"; then
  PHP_ADD_INCLUDE($srcdir)
  PHP_NEW_EXTENSION(nebula_probe, nebula_probe.c hooks.c utils.c, $ext_shared)
fi
