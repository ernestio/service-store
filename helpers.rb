# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

module Sinatra
  module API
    # rubocop:disable Metrics/ModuleLength
    module Helpers
      # Get fake datacenter
      def fake_datacenter
        { datacenter_id: 'fake',
          datacenter_name: 'fake',
          datacenter_username: 'fake',
          datacenter_password: 'fake',
          datacenter_region: 'fake',
          datacenter_type: 'fake',
          external_network: 'fake',
          vse_url: 'http://vse.url/',
          vcloud_url: 'fake' }
      end

      # Get the datacenter by its name
      def get_datacenter(token, name)
        uri = URI.parse("#{ENV['ERNEST_URI']}/datacenters/search?name=#{name}")
        http = Net::HTTP.new(uri.host, uri.port)
        http.use_ssl = true
        http.verify_mode = OpenSSL::SSL::VERIFY_NONE
        request = Net::HTTP::Get.new(uri.request_uri)
        request.initialize_http_header('X-AUTH-TOKEN' => token)
        response = http.request(request)
        return nil if response.code == '404'
        datacenter = JSON.parse(response.body, symbolize_names: true)
        datacenter
      end

      # Get the client by its id
      def get_client(token, id)
        uri = URI.parse("#{ENV['ERNEST_URI']}/clients/#{id}")
        http = Net::HTTP.new(uri.host, uri.port)
        http.use_ssl = true
        http.verify_mode = OpenSSL::SSL::VERIFY_NONE
        request = Net::HTTP::Get.new(uri.request_uri)
        request.initialize_http_header('X-AUTH-TOKEN' => token)
        response = http.request(request)
        return nil if response.code == '404'
        client = JSON.parse(response.body, symbolize_names: true)
        client
      end

      def decode_service(body, ctype)
        # If the content type is YAML convert it to JSON
        if ctype == 'application/json'
          service = JSON.parse(body, symbolize_names: true)
        elsif ctype == 'application/yaml'
          begin
            service = YAML.load(body, symbolize_names: true)
          rescue Psych::SyntaxError
            raise ArgumentError, '0002'
          end
          service = service.each_with_object({}) { |(k, v), memo| memo[k.to_sym] = v }
        else
          fail ArgumentError, '0001'
        end
        service
      end

      def post_service(uri, client, result, id, datacenter, service, previous)
        previous_service_definition = nil
        if result
          if result[:service_status] == 'in_progress'
            halt(400, 'Service is already applying some changes, please wait until they are done')
          end
          previous_service_definition = JSON.parse(result[:service_result]) if result[:service_result]
        end
        data = {
          id:                 id,
          client:             client,
          datacenter:         datacenter,
          service:            service,
          previous:           previous_service_definition
        }
        data[:previous_id] = previous[:service_id] unless previous.nil?

        Net::HTTP.post_form(uri, 'service' => data.to_json)
      end

      def delete_service(uri, _client, result, id)
        if result[:service_status] == 'in_progress'
          halt(400, 'Service is already applying some changes, please wait until they are done')
        end

        req = Net::HTTP::Delete.new("/service/#{id}")
        Net::HTTP.new(uri.host, uri.port).start { |http| http.request(req) }
      end

      def patch_service(uri, id)
        req = Net::HTTP::Patch.new("/service/#{id}")
        Net::HTTP.new(uri.host, uri.port).start { |http| http.request(req) }
      end

      def get_service(uri)
        http = Net::HTTP.new(uri.host, uri.port)
        req = Net::HTTP::Get.new(uri.path)
        http.request(req)
      end
    end
  end
end
