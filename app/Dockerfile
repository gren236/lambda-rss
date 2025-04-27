FROM public.ecr.aws/docker/library/golang:1.24 as build-image
WORKDIR /src
COPY go.mod go.sum main.go ./
RUN go build -o lambda-handler

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build-image /app/lambda-handler .
ENTRYPOINT ./lambda-handler
