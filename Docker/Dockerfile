FROM alpine:latest
COPY twsnmpfc /
RUN  chmod +x /twsnmpfc && mkdir /datastore
EXPOSE 8080/tcp 514/udp 2055/udp 162/udp 
WORKDIR /
ENTRYPOINT [ "/twsnmpfc" ]
