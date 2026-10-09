FROM alpine:3 AS prepare

COPY ./dist/* /usr/local/bin/traefik2traefik

RUN chmod +x /usr/local/bin/traefik2traefik


FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=prepare /usr/local/bin/traefik2traefik /usr/local/bin/traefik2traefik

ENTRYPOINT [ "/usr/local/bin/traefik2traefik" ]
