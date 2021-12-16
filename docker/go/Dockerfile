FROM golang:1.17-alpine

RUN apk update
RUN apk add curl \
    git \
    pkgconfig \
    vips-dev \
    gcc \
    libc-dev
RUN curl https://glide.sh/get | sh

WORKDIR /var/www/cropler
#RUN addgroup --system --gid 1000 www
#RUN adduser -S -u 1000 -g "" www -G www-data -G www
#COPY --chown=www:www . /var/www/cropler
#USER www

# Copy the source from the current directory to the Working Directory inside the container
ADD . /var/www/cropler

RUN go get
RUN go build cropler

CMD ["cropler"]