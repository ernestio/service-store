# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

require 'yaml'
require File.expand_path '../spec_helper.rb', __FILE__

describe 'services_data_microservice' do
  describe 'a non authorized access' do
    describe 'to create service' do
      it 'should throw a 403' do
        post '/services'
        expect(last_response.status).to be 403
      end
    end
    describe 'get services list' do
      it 'should throw a 403' do
        get '/services'
        expect(last_response.status).to be 403
      end
    end
    describe 'get a specific service' do
      it 'should throw a 403' do
        get '/services/foo'
        expect(last_response.status).to be 403
      end
    end
    describe 'update a service' do
      it 'should throw a 403' do
        put '/services/foo'
        expect(last_response.status).to be 403
      end
    end
    describe 'delete a service' do
      it 'should throw a 403' do
        delete '/services/foo'
        expect(last_response.status).to be 403
      end
    end
  end

  describe 'an authorized access' do
    let!(:user_id)    { SecureRandom.uuid }
    let!(:client_id)  { 'client_id' }
    let!(:username)   { 'username' }
    let!(:password)   { 'password' }

    before do
      ServiceModel.dataset.destroy
      @token = SecureRandom.hex
      AuthCache.set @token, { user_id: user_id,
                              client_id: client_id,
                              user_name: username,
                              admin: false }.to_json
      AuthCache.expire @token, 3600
    end

    it 'should get a valid token' do
      expect(@token).to_not be_nil
    end

    describe 'create service' do
      let!(:datacenter_id)             { SecureRandom.uuid }
      let!(:created_service_name)    { 'created_service_name' }
      let!(:created_service_type)    { 'cdg' }
      let!(:created_service_version) { '1.0.0' }
      let!(:created_service_options) { { cdp: { size: 2 } } }
      let!(:created_service) do
        {
          service: 'service',
          datacenter: '',
          name: created_service_name,
          type: created_service_type,
          version: created_service_version,
          options: created_service_options,
          service_status: 'done',
          service_endpoint: 's'
        }
      end

      describe 'for json input' do
        begin
          before do
            VCR.use_cassette(:datacenter_id, record: :new_episodes) do
              post '/services/', created_service.to_json,
                   'HTTP_X_AUTH_TOKEN' => @token, 'CONTENT_TYPE' => 'application/json'
            end
          end
          it 'should respond with a 200 code' do
            expect(last_response.status).to be 200
          end
          it 'should store the user on database' do
            services = ServiceModel.dataset.filter(service_name: created_service_name)
            expect(services.count).to be(1)
          end
        rescue => e
          puts e.backtrace
          raise e
        end
      end

      describe 'for yaml input' do
        let!(:created_service) do
          {
            service: 'service',
            datacenter: '',
            'name' => created_service_name,
            'type' => created_service_type,
            'version' => created_service_version,
            'options' => created_service_options
          }
        end
        before do
          VCR.use_cassette(:datacenter_id) do
            post '/services/', created_service.to_yaml,
                 'HTTP_X_AUTH_TOKEN' => @token, 'CONTENT_TYPE' => 'application/yaml'
          end
        end
        it 'should respond with a 200 code' do
          expect(last_response.status).to be 200
        end
        it 'should store the user on database' do
          services = ServiceModel.dataset.filter(service_name: created_service_name)
          expect(services.count).to be(1)
        end
      end

      describe 'for invalid yaml input' do
        let!(:created_service) do
          {
            'name' => created_service_name,
            'type' => created_service_type,
            'version' => created_service_version,
            'options' => created_service_options
          }
        end
        before do
          VCR.use_cassette(:invalid_datacenter) do
            post '/services/', created_service.to_yaml,
                 'HTTP_X_AUTH_TOKEN' => @token, 'CONTENT_TYPE' => 'application/yaml'
          end
        end
        it { expect(last_response.status).to be 404 }
        it { expect(last_response.body).to eq 'Specified datacenter does not exist' }
      end

      describe 'for invalid content type input' do
        let!(:created_service) do
          {
            'name' => created_service_name,
            'type' => created_service_type,
            'version' => created_service_version,
            'options' => created_service_options
          }
        end
        before do
          VCR.use_cassette(:datacenter_id) do
            post '/services/', created_service.to_yaml,
                 'HTTP_X_AUTH_TOKEN' => @token, 'CONTENT_TYPE' => 'application/hander'
          end
        end
        it { expect(last_response.status).to be 415 }
      end
    end

    describe 'list services' do
      describe 'with existing  records' do
        before do
          1.upto(9) do |i|
            service = ServiceModel.new
            service.service_id = SecureRandom.uuid
            service.datacenter_id = SecureRandom.uuid
            service.service_name = "test_#{i}"
            service.service_type = 'cdg'
            service.service_version = '1.0.0'
            service.service_options = '{"cdp": { "size": 2 }}'
            service.service_definition = '{}'
            service.client_id = client_id
            service.service_status = 'x'
            service.service_endpoint = ''
            service.save
          end
          1.upto(4) do |i|
            service = ServiceModel.new
            service.service_id = SecureRandom.uuid
            service.datacenter_id = SecureRandom.uuid
            service.service_name = 'test_grouped'
            service.service_type = 'cdg'
            service.service_version = i
            service.service_options = '{"cdp": { "size": 2 }}'
            service.service_definition = '{}'
            service.client_id = client_id
            service.service_status = 'x'
            service.service_endpoint = ''
            service.save
          end
          get '/services/', {}, 'HTTP_X_AUTH_TOKEN' => @token
        end
        it 'should response with a 200 code' do
          expect(last_response.status).to be 200
        end
        it 'should return a list of existing services' do
          expect(JSON.parse(last_response.body).length).to be(10)
        end
        it 'should return the highest version of each service name' do
          sw = 0
          puts JSON.parse(last_response.body).length
          JSON.parse(last_response.body).each do |s|
            if s['service_name'] == 'test_grouped'
              sw += 1
              expect(s['service_version']).equal? '4'
            end
          end
          expect(sw).to be 1
        end
      end
      describe 'without existing records' do
        before do
          get '/services/', {}, 'HTTP_X_AUTH_TOKEN' => @token
        end
        it 'should response with a 200 code' do
          expect(last_response.status).to be 200
        end
        it 'should return a list of existing datacenters' do
          expect(JSON.parse(last_response.body).length).to be(0)
        end
      end
      describe 'with existing records with null service_options' do
        before do
          service = ServiceModel.new
          service.service_id = SecureRandom.uuid
          service.datacenter_id = SecureRandom.uuid
          service.service_name = 'test'
          service.service_type = 'cdg'
          service.service_version = '1.0.0'
          service.service_options = '{}'
          service.service_definition = '{}'
          service.client_id = client_id
          service.service_status = 'x'
          service.service_endpoint = ''
          service.save
          get '/services/', {}, 'HTTP_X_AUTH_TOKEN' => @token
        end
        it 'should response with a 200 code' do
          expect(last_response.status).to be 200
        end
        it 'should return a list of existing datacenters' do
          expect(JSON.parse(last_response.body).length).to be(1)
        end
      end
    end

    describe 'get service details' do
      before do
        service = ServiceModel.new
        service.service_id = 'e828da34-c49e-4457-97b0-a31e9e0e4c07'
        service.datacenter_id = 'e828da34-c49e-4457-97b0-a31e9e0e4c07'
        service.service_name = 'test'
        service.service_type = 'cdg'
        service.service_version = '1.0.0'
        service.service_options = '{"cdp": { "size": 2 }}'
        service.service_definition = '{}'
        service.service_status = ''
        service.service_endpoint = ''
        service.client_id = client_id
        service.save
        get '/services/test', {}, 'HTTP_X_AUTH_TOKEN' => @token
      end
      it 'should response with a 200 code' do
        expect(last_response.status).to be 200
      end
      it 'should return service details' do
        service = JSON.parse(last_response.body, symbolize_names: true)
        expect(service[:service_name]).to eq('test')
        expect(service[:service_id]).to eq('e828da34-c49e-4457-97b0-a31e9e0e4c07')
      end
      describe 'a non existing service' do
        before do
          get '/services/non_existing', '', 'HTTP_X_AUTH_TOKEN' => @token
        end
        it 'should return a Not found response' do
          expect(last_response.status).to be(404)
        end
      end
    end

    describe 'update a service' do
      describe 'an existing service' do
        before do
          service = ServiceModel.new
          service.service_id = 'e828da34-c49e-4457-97b0-a31e9e0e4c07'
          service.datacenter_id = 'e828da34-c49e-4457-97b0-a31e9e0e4c07'
          service.service_name = 'test'
          service.service_type = 'cdg'
          service.service_version = '1.0.0'
          service.service_options = '{"cdp": { "size": 2 }}'
          service.service_definition = '{}'
          service.client_id = client_id
          service.service_status = ''
          service.service_endpoint = ''
          service.save
          get '/services/test', {}, 'HTTP_X_AUTH_TOKEN' => @token
          updated_service = JSON.parse(last_response.body, symbolize_names: true)
          updated_service[:service_options][:cdp][:size] = 4
          put '/services/e828da34-c49e-4457-97b0-a31e9e0e4c07', updated_service.to_json, 'HTTP_X_AUTH_TOKEN' => @token
        end
        it 'should return a 200 response code' do
          expect(last_response.status).to be(200)
        end
      end
      describe 'a service that does not exist' do
        before do
          put '/services/non_existing', '', 'HTTP_X_AUTH_TOKEN' => @token
        end
        it 'should return a Not found response' do
          expect(last_response.status).to be(404)
        end
      end
    end

    describe 'delete a service' do
      describe 'delete an existing service' do
        before do
          service = ServiceModel.new
          service.service_id = 'e828da34-c49e-4457-97b0-a31e9e0e4c07'
          service.datacenter_id = 'e828da34-c49e-4457-97b0-a31e9e0e4c07'
          service.service_name = 'test'
          service.service_type = 'cdg'
          service.service_version = '1.0.0'
          service.service_options = '{"cdp": { "size": 2 }}'
          service.service_definition = '{}'
          service.client_id = client_id
          service.service_status = 'x'
          service.service_endpoint = 'x'
          service.service_result = '{"datacenters":{"items":[{"name":"d1"}]}}'
          service.save
          VCR.use_cassette(:delete_service, record: :new_episodes) do
            delete '/services/test', '', 'HTTP_X_AUTH_TOKEN' => @token
          end
        end
        it 'should return a Not Implemented response' do
          expect(last_response.status).to be(200)
        end
      end
      describe 'a non existing service' do
        before do
          VCR.use_cassette(:delete_service) do
            delete '/services/non_existing', '', 'HTTP_X_AUTH_TOKEN' => @token
          end
        end
        it 'should return a Not found response' do
          expect(last_response.status).to be(404)
        end
      end
    end
  end
end
