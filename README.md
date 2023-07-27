
This is the companion repository with some sample code as used in the video. It has working code for uploading and retrieval of movies. _Why not write the rest?_


### Setup

- Requirements
```
Golang v1.17+, Docker, Kubernetes, Helm Chart
```
- Either Run the API locally or push your own docker image and modify the helm chart
`make run`

`make docker-run`

## Post a movie

```
curl -X POST http://localhost:8081/movie -H 'Content-Type: application/json' -d '{
    "title": "Avenger War Cry",
    "year": 2021,
    "cast": ["
      Robert Downey, Jr.,
      Chris Evans,
      Chris Hemsworth,
      Mark Ruffalo
      "],
    "genres": [
      "Action"
    ]
}'
```

## Retrieve a movie

```
curl localhost:8081/movies/year/2021
```

### SQL Serverless Conn PW: `nzXjwvPU5x3q4EjH`

```
mysql.RegisterTLSConfig("tidb", &tls.Config{
  MinVersion: tls.VersionTLS12,
  ServerName: "gateway01.eu-central-1.prod.aws.tidbcloud.com",
})

db, err := sql.Open("mysql", "hLnbrpfVNLTaw8b.root:nzXjwvPU5x3q4EjH@tcp(gateway01.eu-central-1.prod.aws.tidbcloud.com:4000)/test?tls=tidb")
    
```
