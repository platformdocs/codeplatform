FROM node:18-alpine AS frontendBuilder

WORKDIR /frontend
COPY . .
WORKDIR /frontend/app
RUN npm config set registry https://registry.npm.taobao.org
RUN npm install
RUN npm run build

FROM mongo

FROM golang:1.22.0-alpine AS backendBuilder

WORKDIR /codeplatform
COPY --from=frontendBuilder /frontend .
COPY . .

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod tidy

RUN go build -o main main.go

EXPOSE 8080
CMD ["/codeplatform/main"]