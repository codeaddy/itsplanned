test: 
	@bash tests/setup_test_db.sh
	@go test ./tests/... -v
	@docker stop test_postgres_db