generate_swagger:
	swag init --generalInfo web/router.go --dir ./web --swagger web/docs/swagger/

generate_mocks:
	@echo "Building mocks"
	mockgen -package=manager -source=updater/interfaces.go UpdateProgress,UpdatePlan > updater/manager/updateProgressMock_test.go
	sed 's%x "."%x "kubernetes-update-manager/updater"%g' updater/manager/updateProgressMock_test.go > updater/manager/updateProgressMock_test_sed.go
	mv updater/manager/updateProgressMock_test_sed.go updater/manager/updateProgressMock_test.go
	mockgen -package=updater -source=updater/interfaces.go MatchConfig > updater/matcherMock_test.go

cloc:
	cloc --not-match-f="(cloc.xml|swagger.*|cover.out|coverage.xml|xunit.xml)" --exclude-d vendor .

lint:
	bash ./scripts/run-golint.sh

xunit:
	bash ./scripts/run-xunit-tests.sh
