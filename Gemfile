source 'https://rubygems.org'

gem 'flowauth', path: '/opt/ernest-libraries/authentication-middleware'
gem 'nats', git: 'https://github.com/r3labs/ruby-nats.git'

gem 'sinatra'
gem 'pg'
gem 'sequel'

group :development, :test do
  gem 'pry'
end

group :test do
  gem 'rspec'
  gem 'rack-test'
  gem 'rubocop',   require: false
  gem 'simplecov', require: false
  gem 'vcr'
  gem 'webmock'
end
