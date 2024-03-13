FROM golang:bookworm
RUN  apt update && apt upgrade -yq && apt install -yq  \
gcc-multilib gcc-mingw-w64 binutils-mingw-w64
RUN  apt install -yq g++-aarch64-linux-gnu g++-arm-linux-gnueabihf binutils-arm-linux-gnueabihf
COPY mk.sh /root/
RUN  chmod +x /root/mk.sh
ENTRYPOINT [ "/root/mk.sh" ]
 