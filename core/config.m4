PHP_ARG_ENABLE(flownebula,
  whether to enable FlowNebula profiler,
  [  --enable-flownebula       Enable FlowNebula profiler])

if test "$PHP_FLOWNEBULA" != "no"; then
  PHP_NEW_EXTENSION(flownebula, flownebula.c, $ext_shared)
fi