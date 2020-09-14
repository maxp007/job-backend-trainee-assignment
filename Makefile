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
	  $(RM) bill_service$(EXT)
	  $(RM) app_log.log
	  $(RM) cover.out
	  $(RM) cover.html
gen_doc:
	swagger generate spec /o ./docs/swagger.json /m

serve_doc:
	 swagger serve -F=swagger ./docs/swagger.json

tests:
	go test -race ./...