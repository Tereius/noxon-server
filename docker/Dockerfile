FROM --platform=$BUILDPLATFORM golang:1.21 as build

WORKDIR /build
COPY ./ ./
RUN go mod download

ARG UID=11009 GID=11009
RUN echo "noxon:*:$UID:$GID::/:" > passwd && echo "noxon:*:$GID:" > group

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -a -tags netgo,osusergo -trimpath -ldflags '-w -s -buildid= -extldflags "-static"' -o noxon-server cmd/main.go


FROM scratch as run

WORKDIR /noxon
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/passwd /build/group /etc/
COPY --from=build /build/noxon-server ./
ENV GIN_MODE=release
USER noxon:noxon
EXPOSE 80/tcp 53/udp
ENTRYPOINT ["/noxon/noxon-server"]
