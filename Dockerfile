FROM ubuntu
COPY ./main /main
ENTRYPOINT ["/main"]
EXPOSE 8080
