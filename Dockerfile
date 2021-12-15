FROM golang:1.17-alpine

RUN apk update && apk add curl git pkgconfig vips-dev gcc libc-dev && curl https://glide.sh/get | sh

WORKDIR /var/www/cropler
# Copy the source from the current directory to the Working Directory inside the container
ADD . /var/www/cropler

RUN go get

RUN go build cropler

CMD ["cropler"]