default: install

lint:
	bundle exec rubocop

install:
	bundle install

cover:
	COVERAGE=true MIN_COVERAGE=0 bundle exec rspec -c -f d spec

test:
	bundle exec rspec
