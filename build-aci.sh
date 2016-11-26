#!/bin/sh

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export VERSION=v0.1.1

go build -o bin/rkt-sidekick -ldflags '-extldflags "-static"'
acbuild begin
acbuild set-name monder.cc/rkt-sidekick
acbuild copy bin/rkt-sidekick /bin/rkt-sidekick
acbuild set-exec /bin/rkt-sidekick
acbuild label add version $VERSION
acbuild label add arch $GOARCH
acbuild label add os $GOOS
acbuild annotation add authors "Aleksejs Sinicins <monder@monder.cc>"
acbuild write bin/rkt-sidekick-${VERSION}-${GOOS}-${GOARCH}.aci --overwrite
acbuild end
