FROM golang
WORKDIR /app
RUN go install github.com/popcell/chatgpt-proxy@latest
CMD [ "chatgpt-proxy" ]