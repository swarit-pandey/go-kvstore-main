# Golang base image 
FROM golang:1.17-alpine as builder 

# Tools 
RUN apk update && apk add --no-cache git 

# Set the current working directory 
WORKDIR /app

# Copy dependencies 
COPY go.mod go.sum ./

# To download the dependencies 
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/go-kvstore

# Reduce final image size 
FROM alpine:latest 

# Set the current working directory inside container
WORKDIR /app 

# Copy the binary from the previous stage
COPY --from=builder /app/main . 

# Expose the port the application will run on 
EXPOSE 8080 

# Run the application 
CMD ["./main"] 