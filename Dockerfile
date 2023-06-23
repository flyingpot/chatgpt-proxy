FROM golang
RUN go install github.com/flyingpot/chatgpt-proxy@latest
CMD [ "chatgpt-proxy" ]
