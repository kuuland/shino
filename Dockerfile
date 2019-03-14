FROM alpine
RUN rm -rf /etc/apk/repositories && \
  echo "http://mirrors.aliyun.com/alpine/v3.8/main" > /etc/apk/repositories
ADD ./dist/shino /bin/
RUN apk -Uuv add ca-certificates
ENTRYPOINT /bin/shino