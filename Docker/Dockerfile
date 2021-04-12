FROM alpine:latest
RUN apk add --update --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    echo "Asia/Tokyo" > /etc/timezone && \
    apk del tzdata
COPY twsnmpfc /
RUN  chmod +x /twsnmpfc && mkdir /datastore
EXPOSE 8080/tcp 514/tcp 514/udp 2055/udp 162/udp 
WORKDIR /
ENTRYPOINT [ "/twsnmpfc" ]
