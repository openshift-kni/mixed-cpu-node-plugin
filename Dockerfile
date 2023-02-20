FROM golang
COPY build/bin/mixedcpus /bin/mixedcpus
ENTRYPOINT [ "/bin/mixedcpus" ]