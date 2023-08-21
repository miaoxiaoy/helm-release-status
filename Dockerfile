FROM golang as builder
WORKDIR /code
ENV GOOS=linux
ENV CGO_ENABLED=0
COPY . /code
RUN go build -o /app/helm-release-status


FROM alpine:3
RUN apk add aws-cli
RUN apk add curl
COPY --from=builder /app/helm-release-status /helm-release-status
CMD [ "/helm-release-status" ]
