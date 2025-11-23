.PHONY: gen up down clean 

gen:                               
	oapi-codegen -package api -o internal/api/api.gen.go openapi.yaml

up:                                      
	docker-compose up --build

down:                              
	docker-compose down

clean:                                   
	rm -f internal/api/openapi.gen.go