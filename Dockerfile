FROM gcr.io/distroless/static-debian13:nonroot

COPY ./dist/* /traefik2traefik

USER 1000:1000

ENTRYPOINT [ "/traefik2traefik" ]
