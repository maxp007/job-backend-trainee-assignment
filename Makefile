target: all

ifdef OS
   RM = del
   EXT = .exe
else
   ifeq ($(shell uname), Linux)
      RM = rm
   endif
endif

all:
	go build -o bill_service$(EXT) -mod=vendor

clean:
	  $(RM)  -r ./log bill_service$(EXT) app_log.log cover.out cover.html

gen_doc:
	swagger generate spec /o ./docs/swagger.json /m

serve_doc:
	 swagger serve -F=swagger ./docs/swagger.json

test:
	go test -race ./...  && \
	cd ./internal/testing_dockerfiles/app_testing &&  docker-compose up --build --abort-on-container-exit --exit-code-from testing_app &&\
	cd ../../../ && \
	cd ./internal/testing_dockerfiles/http_handler_testing &&  docker-compose up --build --abort-on-container-exit --exit-code-from testing_app &&\
	cd ../../../ && \
    cd ./internal/testing_dockerfiles/cache_testing &&  docker-compose up --build --abort-on-container-exit --exit-code-from testing_app