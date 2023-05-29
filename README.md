# Fortress API

<p align="center">
  <img src="https://img.shields.io/badge/golang-1.18-blue" />
  <img src="https://img.shields.io/badge/strategy-gitflow-%23561D25" />
  <a href="https://github.com/consolelabs/mochi-api/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-GNU-blue" />
  </a>
</p>

## Overview

This repository is the official BE service for Fortress

## How to contribute

### Prerequisites

1. Go installed
2. Docker installed

### How to run source code locally

1. Set up source

Set up infras, install dependencies, etc.

```
make init
```

2. Set up env

Create a file `.env` with these values:

```
DB_HOST="127.0.0.1"
DB_PORT="25432"
DB_USER="postgres"
DB_PASS="postgres"
DB_NAME="fortress_local"
DB_SSL_MODE="disable"
ALLOWED_ORIGINS="*"
ENV=dev
DEBUG=true
JWT_SECRET_KEY=JWTSecretKey
```

3. Run source

```
make dev
```

The service starts with port 8080 as the default

### How to work on a TODO

1. Feel free to pick any TODO for you from [Board View](https://www.notion.so/dwarves/4d756d46e90240918cd2505f962cacd1?v=d65335d1772f4532ab1bc274a1ae8c76)
2. **Assign** that item to your account
3. Remember to update item’s **status** based on your working progress
   - `Not Started`: not started yet
   - `Planned`: working on this sprint
   - `In Progress`: still working on
   - `In Review`: Task done = PR has been merged to `develop` branch at least
   - `Complete`: Confirmation from the team that the TODO is finished

### PR template

```markdown
#### What's this PR do?

- [x] Add new routes for user
- [x] Add instruction for using docker-compose

#### What are the relevant Git tickets?

// Put in link to Git Issue

#### Screenshots (if appropriate)

// Use [Licecap](http://www.cockos.com/licecap/) to share a screencast gif.

#### Any background context you want to provide? (if appropriate)

- Is there a blog post?
- Does the knowledge base need an update?
- Does this add new dependencies which need to be added to?
```

## Technical Document

### Project structure

- `cmd/` this folder contains the main application entry point files for the project
- `docs/`: contains Swagger documentation files generated by [swaggo](https://github.com/swaggo/swag)
- `migrations/`: contains seeds and SQL migration files
  - `schemas/`: contains DB schema migration files
  - `seed/`: contains seed files which will initialize DB with sets of dummy data
  - `test-seed/`: also seed files but for test DB
- `pkg/`: contains core source code of service
  - `config/`: contains configs for application
  - `handler/`: handling API requests
  - `logger/`: logging initial and functional methods
  - `model/`: DB model structs
  - `mw/`: middleware
  - `request/`: API request models
  - `routes/`: API routing (see [gin](https://github.com/gin-gonic/gin))
  - `service/`: contains interaction with external services (google API, etc.)
  - `store/`: data access layers, contains DB CRUD operations (see [gorm](https://gorm.io/))
  - `utils/`: utility methods
  - `view/`: API view models

### Sample usecases

   1. Create new API
      - Check out file `/pkg/routes/v1.go` and explore the code flow to see how to create and handle an API
      - Remember to annotate handler functions with [swaggo](https://github.com/swaggo/swag). Then run `make gen-swagger` to generate Swagger documentations

   2. New DB migration

      Check out `.sql` files under `/migrations` to write a valid schema migration / seed file

      - To apply new migration files, run `make migrate-up`
      - To apply seed files, run `make seed-db`
      - To apply new migration files for test DB, run `make migrate-test`

      **Note:** remember to run these 2 every time you pulling new code

      ```
      make migrate-up
      make migrate-test
      ```

   3. DB repositories

      Check out dirs under `/pkg/store`

## :pray: Credits

A big thank to all who contributed to this project!

If you'd like to contribute, please contact us.

