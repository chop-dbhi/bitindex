FROM scratch

MAINTAINER Byron Ruth <b@devel.io>

COPY ./dist/bitindex-linux-amd64 /main

EXPOSE 7000

ENTRYPOINT ["/main"]

VOLUME ["/data.bitx"]

CMD ["http", "--host=0.0.0.0", "/data.bitx"]
