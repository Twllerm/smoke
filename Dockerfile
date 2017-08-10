FROM golang


COPY . /go/src/smoke3/
COPY ./lib/pq /go/src/github.com/lib/pq/
COPY ./lib/bot /go/src/bot
WORKDIR /go/src/smoke3
RUN go install
CMD ["smoke3", "--runDDL=true"]