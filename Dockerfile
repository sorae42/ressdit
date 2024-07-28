FROM golang:1.22 as builder

WORKDIR /app

COPY go.* ./
RUN mkdir pkg
COPY pkg/reddit ./pkg/reddit
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server ./cmd/ressdit
FROM alpine:edge

COPY --from=builder /app/server /server
EXPOSE 8080

ENV PORT="8080"
ENV REDDIT_URL="https://www.reddit.com"

CMD ["/server"]
