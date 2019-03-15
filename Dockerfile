FROM node:carbon-alpine
ADD shino /bin/
RUN apk -Uuv add ca-certificates git
ENTRYPOINT shino