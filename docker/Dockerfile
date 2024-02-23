FROM node:18 AS frontendBuilder

# set app work dir
WORKDIR /codeplatform

# copy assets files to the container
COPY app/ .

# set assets/ as work dir to build frontend static files
WORKDIR /codeplatform/app
RUN npm install
RUN npm run build

FROM golang:1.16.3 AS backendBuilder

# set app work dir
WORKDIR /codeplatform

# copy all files to the container
COPY main.go .

# start app
CMD [ "go", "run", "main.go" ]
