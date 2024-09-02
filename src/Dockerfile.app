FROM --platform=$BUILDPLATFORM golang:alpine AS setup

ENV GO111MODULE=on

# Copy all the module and workspace files
WORKDIR /app
COPY ./app/go.mod ./app/
COPY ./app/go.sum ./app/

COPY ./shared/go.mod ./shared/
COPY ./shared/go.sum ./shared/

# Initialize the workspace for the project
RUN go work init app shared

# Download all the dependencies
WORKDIR /app/app
RUN go mod download

WORKDIR /app/shared
RUN go mod download

# Copy the source code and run the code
WORKDIR /app
COPY ./app ./app
COPY ./shared ./shared


# Install and run the code using the dev server
FROM setup AS dev

WORKDIR /app/app
RUN go install github.com/gravityblast/fresh@8d1fef547a99be2395e7587f8de5d01265176650

CMD ["fresh"]


# Compile the code
FROM setup AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app/app
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build


# Run the compiled code
FROM alpine:latest AS run

# TODO Set to release mode
#ENV GIN_MODE=release
COPY --from=build /app/app/app /app
ENTRYPOINT ["/app"]