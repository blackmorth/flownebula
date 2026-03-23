FROM php:8.3-fpm

RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    autoconf \
    automake \
    libtool \
    pkg-config \
    procps \
    && rm -rf /var/lib/apt/lists/*

# Copier ton extension
COPY probe /usr/src/nebula_probe

WORKDIR /usr/src/nebula_probe

# Compilation du module PHP
RUN phpize
RUN ./configure
RUN make && make install

# Charger l’extension
RUN echo "extension=nebula_probe.so" > /usr/local/etc/php/conf.d/nebula.ini \
    && echo "nebula_probe.enabled=1" >> /usr/local/etc/php/conf.d/nebula.ini

WORKDIR /var/www/html

COPY test.php .

CMD ["php-fpm"]
