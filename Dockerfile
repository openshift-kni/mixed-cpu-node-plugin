FROM golang:1.20
COPY build/bin/mixedcpu /bin/mixedcpu
ENTRYPOINT [ "/bin/mixedcpu" ]
