go list ./... | grep -v /vendor/ | xargs golint > golint.xml
go vet ./... > govet.xml
