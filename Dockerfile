FROM php:8.2-cli

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
        autoconf \
        build-essential \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY core /app/core

RUN cd /app/core \
    && phpize \
    && ./configure --enable-flownebula \
    && make \
    && make install

RUN echo "extension=flownebula.so" >> /usr/local/etc/php/conf.d/flownebula.ini \
    && echo "flownebula.trace_path=/tmp/nebula.trace" >> /usr/local/etc/php/conf.d/flownebula.ini

COPY analyzer /app/analyzer
COPY viewer /app/viewer
COPY examples /app/examples

EXPOSE 8080

# Serve from /app so that /viewer/ and /nebula.json are both available
CMD ["php", "-S", "0.0.0.0:8080", "-t", "/app"]

