# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

require 'sinatra'
require 'sequel'
require 'flowauth'
require 'yaml'
require 'net/http'
require 'securerandom'
require 'nats/client'
require 'uri'

require_relative 'helpers'

module Config
  def self.load_db
    NATS.start(servers: [ENV['NATS_URI']]) do
      NATS.request('config.get.postgres') do |r|
        return JSON.parse(r, symbolize_names: true)
      end
    end
  end
  def self.load_redis
    return if ENV['RACK_ENV'] == 'test'
    NATS.start(servers: [ENV['NATS_URI']]) do
      NATS.request('config.get.redis') do |r|
        return r
      end
    end
  end
end

class API < Sinatra::Base
  helpers Sinatra::API::Helpers
  configure do
    # Default DB Name
    ENV['DB_URI'] ||= Config.load_db[:url]
    ENV['DB_REDIS'] ||= Config.load_redis
    ENV['DB_NAME'] ||= 'services'

    # TODO: depreciate http calls in favor of NATS
    ENV['ERNEST_URI'] ||= 'https://ernest.local'
    ENV['GPB_SERVICE_URL'] ||= 'http://127.0.0.1:21000'

    #  Initialize database
    Sequel::Model.plugin(:schema)
    DB = Sequel.connect("#{ENV['DB_URI']}/#{ENV['DB_NAME']}")

    # Create services table if not exists
    DB.create_table? :services do
      String :service_id, null: false, primary_key: true
      String :client_id, null: false
      String :datacenter_id, null: false
      String :service_name, null: false
      String :service_type, null: false
      String :service_version, null: false
      String :service_status, null: false
      Text :service_options, null: false
      Text :service_definition
      Text :service_result
      Text :service_error
      String :service_endpoint, null: false
    end

    Object.const_set('ServiceModel', Class.new(Sequel::Model(:services)))
  end

  # Set content type as a JSON for all requests
  before do
    content_type :json
  end

  # All requests are autenticated
  use Authentication

  # POST /services/generate
  post '/services/uuid/?' do
    payload = JSON.parse(request.body.read)
    uuid = Digest::MD5.hexdigest(payload['id'])

    status 200
    return { uuid: uuid }.to_json
  end

  # POST /services
  #
  # Create a service
  post '/services/?' do
    begin
      # Get the payload Content-Type
      body = request.body.read
      ctype = request.env['CONTENT_TYPE']
      service = decode_service(body, ctype)

      # Check the service name
      halt 400, 'Service name can\'t be null' if service[:name].nil?

      # Get the datacenter data
      datacenter = fake_datacenter
      if service[:provider] != 'fake'
        datacenter = get_datacenter(env['HTTP_X_AUTH_TOKEN'], service[:datacenter])
      end
      halt(404, 'Specified datacenter does not exist') if datacenter.nil?

      # Get the client data
      client = get_client(env['HTTP_X_AUTH_TOKEN'], env[:current_user][:client_id])

      # Get previous execution if exists
      result = ServiceModel.filter(service_name: service[:name], client_id: env[:current_user][:client_id]).order(Sequel.desc(:service_version)).first

      # Send service to gpb-service-creator-microservice
      uri = URI("#{ENV['GPB_SERVICE_URL']}/service")
      id = Digest::MD5.hexdigest(service[:name] + '-' + service[:datacenter])
      id = "#{SecureRandom.uuid}-#{id}"

      if !result.nil? && !result[:service_error].nil?
        res = patch_service(uri, result[:service_id])
        id = result[:service_id]
      else
        begin
          res = post_service(uri, client, result, id, datacenter, service, result)
        rescue EOFError => e
          halt 400, 'Provided yaml is not valid'
        end
        fail(ArgumentError, res.body) if res.code.to_s != '200'

        ServiceModel.insert(service_id: id,
                            service_name: service[:name],
                            datacenter_id: datacenter[:datacenter_id],
                            client_id: env[:current_user][:client_id],
                            service_version: Time.now.to_i,
                            service_type: 'vcloud',
                            service_options: '{}',
                            service_status: 'in_progress',
                            service_endpoint: '',
                            service_definition: body)
      end

      status 200
      return { id: id }.to_json

    rescue ArgumentError => e
      if e.message == '0001'
        halt 415, 'Unsupported Media Type. Supported media types are application/json and application/yaml'
      elsif e.message == '0002'
        halt 400, 'Yaml is invalid'
      elsif e.message != ''
        halt 400, e.message
      end
    rescue => e
      puts e
      puts e.backtrace
    end
  end

  # GET /services
  #
  # Fetch all services
  get '/services/?' do
    filters = { client_id: env[:current_user][:client_id] }
    data = ServiceModel.filter(filters).order(Sequel.asc(:service_name), Sequel.desc(:service_version)).all

    services = []
    name = ''
    data.each do |row|
      if name != row[:service_name]
        name = row[:service_name]
        services.push(
          service_id: row[:service_id],
          datacenter_id: row[:datacenter_id],
          service_name: row[:service_name],
          service_version: row[:service_version],
          service_status: row[:service_status],
          service_options: row[:service_options],
          service_endpoint: row[:service_endpoint]
        )
      end
    end
    services.to_json
  end

  # GET /services/search
  #
  # Search an service by its properties
  get '/services/search/?' do
    filters = { client_id: env[:current_user][:client_id] }
    filters[:service_name] = params[:name] if params.include? :name
    filters[:datacenter_id] = params[:datacenter] if params.include? :datacenter

    service = ServiceModel.filter(filters).order(Sequel.desc(:service_version)).first
    halt 404 if service.nil?
    status 200
    return { service_id:       service[:service_id],
             datacenter_id:    service[:datacenter_id],
             service_name:     service[:service_name],
             service_version:  service[:service_version],
             service_status: service[:service_status],
             service_options:  JSON.parse(service[:service_options]),
             service_endpoint: service[:service_endpoint],
             service_definition: service[:service_definition],
             service_result: service[:service_result] }.to_json
  end

  # GET /services/:service
  #
  # Fetch an service by its ID
  get '/services/:service/?' do
    service = ServiceModel.filter(client_id: env[:current_user][:client_id], service_name: params[:service]).order(Sequel.desc(:service_version)).first
    halt 404 if service.nil?
    status 200
    return { service_id:       service[:service_id],
             datacenter_id:    service[:datacenter_id],
             service_name:     service[:service_name],
             service_version:  service[:service_version],
             service_status: service[:service_status],
             service_options:  JSON.parse(service[:service_options]),
             service_endpoint: service[:service_endpoint],
             service_definition: service[:service_definition],
             service_result: service[:service_result] }.to_json
  end

  # PUT /services/:service
  #
  # Updates an service by its ID
  put '/services/:service/?' do
    begin
      service = ServiceModel.filter(client_id: env[:current_user][:client_id], service_id: params[:service]).first
      halt 404 if service.nil?
      status 200
      updated_service = JSON.parse(request.body.read, symbolize_names: true)
      updated_service[:client_id] = env[:current_user][:client_id]
      updated_service[:service_id] = service[:service_id]
      updated_service[:service_definition] = service[:service_definition]
      updated_service[:service_endpoint] = service[:service_endpoint]
      service.update(service_definition: service[:service_definition], service_endpoint: service[:service_endpoint])
    rescue => e
      puts e
      puts e.backtrace
    end
  end

  # POST /services/:service/reset
  #
  # Resets a service by its Name
  post '/services/:service/reset/?' do
    service = ServiceModel.filter(client_id: env[:current_user][:client_id], service_name: params[:service]).order(Sequel.desc(:service_version)).first
    halt 404, "No services found with for '#{params[:service]}'" if service.nil?
    halt 404, "Reset only applies to 'in progress' serices, however service '#{params[:service]}' is on status '#{service[:service_status]}'" if service[:service_status] != 'in_progress'

    uri = URI.parse("#{ENV['GPB_SERVICE_URL']}/service/#{service[:service_id]}")
    res = get_service(uri)
    service.update(service_status: 'errored', service_error: res.body)
    status 200
  end

  # DELETE /services/:services
  #
  # Deletes an service by its Name
  delete '/services/:service/?' do
    begin
      # Get the payload Content-Type
      client = get_client(env['HTTP_X_AUTH_TOKEN'], env[:current_user][:client_id])

      # Get previous execution if exists
      result = ServiceModel.filter(service_name: params[:service], client_id: env[:current_user][:client_id]).order(Sequel.desc(:service_version)).first
      if !result.nil?
        uri = URI("#{ENV['GPB_SERVICE_URL']}/service")
        delete_service(uri, client, result, result[:service_id])
        status 200
        stream = result[:service_id].split('-').last
        return { id: result[:service_id], stream_id: stream }.to_json
      elsif result.nil?
        halt 404, "Service '#{params[:service]}' not found"
      end
    rescue => e
      puts e
      puts e.backtrace
      halt 500, 'An error ocurred'
    end
  end

  # GET /services/:services/builds
  #
  # Fetch all builds for a service by its name
  get '/services/:service/builds/?' do
    filters = { client_id: env[:current_user][:client_id], service_name: params[:service] }
    data = ServiceModel.filter(filters).order(Sequel.desc(:service_version)).all
    services = []
    data.each do |row|
      services.push(
        service_id: row[:service_id],
        datacenter_id: row[:datacenter_id],
        service_name: row[:service_name],
        service_version: row[:service_version],
        service_status: row[:service_status],
        service_options: row[:service_options],
        service_endpoint: row[:service_endpoint]
      )
    end
    services.to_json
  end

  # GET /services/:services/builds/:build
  #
  #
  # Fetch a build from a service by its Name
  get '/services/:service/builds/:build/?' do
    filters = { client_id: env[:current_user][:client_id], service_name: params[:service], service_id: params[:build] }
    data = ServiceModel.filter(filters).order(Sequel.desc(:service_version)).first
    halt 404 if data.nil?
    status 200
    service = {
      service_id: data[:service_id],
      datacenter_id: data[:datacenter_id],
      service_name: data[:service_name],
      service_version: data[:service_version],
      service_status: data[:service_status],
      service_options: data[:service_options],
      service_endpoint: data[:service_endpoint],
      service_definition: data[:service_definition]
    }
    service.to_json
  end
end
