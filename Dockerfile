FROM golang
COPY build/bin/mixedcpu /bin/mixedcpu
ENTRYPOINT [ "/bin/mixedcpu" ]