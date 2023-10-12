#!/bin/bash

case "$IMGPROXY_MALLOC" in

  malloc)
    # Do nothing
    ;;

  jemalloc)
    export LD_PRELOAD="$LD_PRELOAD:/usr/local/lib/libjemalloc.so"
    ;;

  tcmalloc)
    export LD_PRELOAD="$LD_PRELOAD:/usr/local/lib/libtcmalloc_minimal.so"
    ;;

  *)
    echo "Unknows malloc: $IMGPROXY_MALLOC"
    exit 1
esac

exec "$@"
