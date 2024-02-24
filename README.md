# Codeplatform V1.0

##A Golang online code execution platform


- React
- Semantic UI
- Mongo
- Golang
- Gin

### React

For the front end, the follow the steps below is to run in the development environment.

```bash
cd app
npm install
npm run dev
```

Since in the production environment, the server will directly load static files, just run the following code to generate dist folder.

```bash
npm run build
```

### Mongo
The database is a MongoDB instance, you need to install it first. 

```bash
docker pull mongo:latest
```

Then create and start the mongo database instance

```bash
docker run --name codeplatform-mongodb -p 27017:27017 -v [YOUR_DB_DATA_PATH]:[YOUR_DB_DATA_PATH] -d mongo
```

### Golang
Using version : 1.22.0, you need to install it first.
If local version is different, can modify the go.mod file.

And then run the following command to install the dependent library.

```bash
go mod download github.com/gin-gonic/gin
```

Also need to rebuid go.mod file

```bash
go mod tidy
```

to start the server

```bash
go run main.go
```

### Docker
Docker will build the frontend react code and backend golang code, also will start the mongo database instance.
You can run the following command to start the docker container.

```bash
docker-compose up -d
```

You can visit the platform at [CodePlatform](http://localhost:8080)