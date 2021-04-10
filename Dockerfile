FROM ubuntu:bionic as configure
    RUN apt-get update
    RUN apt-get install -y curl gnupg2 lsb-release
    RUN curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
    RUN echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/migrate.list
    RUN apt-get update
    RUN apt-get install -y migrate
    RUN migrate --version

    RUN apt install -y postgresql-client
    RUN psql --version

    WORKDIR /

