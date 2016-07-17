#!/bin/sh

CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
VERSION=0.0.1

go build -ldflags '-extldflags "-static"'
acbuild begin
acbuild set-name monder.cc/rkt-sidekick
acbuild copy rkt-sidekick /bin/rkt-sidekick
acbuild set-exec /bin/rkt-sidekick
acbuild label add version $VERSION
acbuild label add arch $GOARCH
acbuild label add os $GOOS
acbuild annotation add authors "Aleksejs Sinicins <monder@monder.cc>"
acbuild write rkt-sidekick-${VERSION}-${GOOS}-${GOARCH}.aci
acbuild end
