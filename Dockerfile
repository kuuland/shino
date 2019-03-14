FROM alpine
ADD ./dist/shino /bin/
RUN apk -Uuv add ca-certificates
ENTRYPOINT /bin/shino