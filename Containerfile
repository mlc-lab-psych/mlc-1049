FROM docker.io/golang:1.26.2-alpine AS build

WORKDIR /src/
RUN apk add git
COPY go.* .
RUN go mod download

COPY . .

COPY *.go .
RUN go build -v -o experiment

FROM docker.io/alpine
RUN apk add --no-cache tzdata

COPY --from=build /src/templates /template
COPY --from=build /src/experiment /experiment

ENTRYPOINT [ "/experiment" ]