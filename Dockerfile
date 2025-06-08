FROM golang:1.23

LABEL dev.containers.feature="true"

WORKDIR /app

RUN apt-get update && apt-get install -y git curl

CMD [ "sleep", "infinity" ]
