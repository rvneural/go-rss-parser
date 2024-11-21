FROM debian:latest

LABEL maintainer="gafarov@realnoevremya.ru"

RUN apt-get update -y && apt-get upgrade -y
RUN apt-get install -y ca-certificates

EXPOSE 7001

COPY . .

# Максимальное кол-во секунд ожидания загрузки RSS фида
ENV MAX_TIMEOUT=2 

# Кол-во минут, как часто нужно обновлять список фидов в кэше
ENV UPDATE_TIME=30

WORKDIR /build/linux

CMD [ "./rss" ]

