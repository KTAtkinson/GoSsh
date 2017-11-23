from ubuntu:latest
copy GoSsh /bin/
expose 22
entrypoint GoSsh
