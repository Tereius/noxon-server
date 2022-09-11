FROM golang:1.19

WORKDIR /noxon

COPY "./" "./"

RUN useradd -u 10001 noxon
RUN GOOS="linux" && go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o noxon-server cmd/main.go

FROM scratch

ENV GIN_MODE=release
WORKDIR /noxon
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /noxon/noxon-server ./
USER noxon
EXPOSE 80/tcp
EXPOSE 53/udp
ENTRYPOINT ["/noxon/noxon-server"]