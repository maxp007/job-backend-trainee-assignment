### job-backend-trainee-assignment
##### Billing app

### Api Documentation [SwaggerHub page](https://app.swaggerhub.com/apis-docs/maxp007/api_job_backend_trainee_assignment/1.0.0)

## Технологии
* Go 1.14
* PostgreSQL 
* Docker-Compose

## Результаты
* сделаны базовые и дополнительные задачи
* написаны unit и integration тесты    

### Запуск приложения 
    docker-compose up  

### Запуск unit-тестов
    go test -race ./...
      
### Запуск integration-тестов
    cd ./internal/testing_dockerfiles/app_testing &&  docker-compose up --abort-on-container-exit --exit-code-from testing_app
    cd ./internal/testing_dockerfiles/http_handler_testing &&  docker-compose up --abort-on-container-exit --exit-code-from testing_app