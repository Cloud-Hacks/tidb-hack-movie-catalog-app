
Movie Catalogue API is an application programming interface (API) that provides functionality to manage and retrieve information about movies in a movie catalog. It serves as a backend service that allows developers to interact with the movie database, perform CRUD (Create, Read, Update, Delete) operations, and search for movies based on various criteria.


### Setup

- Requirements
```
Golang v1.17+, Docker, Kubernetes, Helm Chart
```
- Either Run the API locally or push your own docker image and modify the helm chart

`make run`

OR

`make docker-start`

## Post a movie

```
curl -X POST http://localhost:8081/movie -H 'Content-Type: application/json' -d '{
    "title": "Avenger War Cry",
    "year": 2021,
    "cast": ["Robert Downey, Jr., Chris Evans, Chris Hemsworth, Mark Ruffalo"],
    "genres": ["Action"]
}'
```

## Retrieve a movie

```
curl localhost:8081/movies/year/2021
```
