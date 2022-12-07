#!/bin/bash
TEMP=$(mktemp -d)
SRC="pcre-8.45"
echo "Using temp directory $TEMP to build $SRC"
(
  cd "$TEMP"
  curl -LO "https://sourceforge.net/projects/pcre/files/pcre/8.45/$SRC.tar.gz"
  tar -xf "$SRC.tar.gz"
  (
    cd "$SRC"
    ./configure \
      --enable-jit \
      --enable-utf \
      --disable-shared \
      --disable-cpp \
      --enable-newline-is-any \
      --with-match-limit=500000 \
      --with-match-limit-recursion=50000
    make -j$(nproc)
  )
)
PLATFORM="$(uname -s)"
ARCH="$(uname -m)"
case "${ARCH}" in
  i*86)               OUTARCH=386;;
  x86_64)             OUTARCH=amd64;;
  arm64 | aarch64)    OUTARCH=arm64;;
  arm*)               OUTARCH=armhf;;
esac
case "${PLATFORM}" in
  Linux*)  OUTPUT=libpcre_linux_${OUTARCH}.a;;
  Darwin*) OUTPUT=libpcre_darwin_${OUTARCH}.a;;
  MINGW*)  OUTPUT=libpcre_windows_${OUTARCH}.a;;
  *)       OUTPUT=libpcre_${OUTARCH}.a
esac
cp "$TEMP/$SRC/.libs/libpcre.a" "$OUTPUT"
echo "Copied static library to $OUTPUT"
rm -rf "$TEMP"
