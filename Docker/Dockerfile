FROM alpine:latest as builder
RUN apk add --update --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    echo "Asia/Tokyo" > /etc/timezone
COPY twsnmpfc /
RUN  chmod +x /twsnmpfc && mkdir /datastore

FROM alpine:latest
COPY --from=builder /twsnmpfc /
COPY --from=builder /etc/timezone /etc/
COPY --from=builder /etc/localtime /etc/
COPY --from=builder /datastore/ /datastore/
EXPOSE 8080/tcp 514/tcp 514/udp 2055/udp 162/udp 
WORKDIR /
ENTRYPOINT [ "/twsnmpfc" ]

