#!/usr/bin/env ruby
# Wraps nmap and converts XML output to JSON

require 'rexml/document'
require 'json'

def scan_network(cidr)
  # Build nmap command
  # -sn: Host discovery only (no port scan) - faster
  # -T4: Aggressive timing template
  # -oX -: Output XML to stdout (- means stdout)
  cmd = "nmap -sn -T4 -oX - #{cidr}"

  xml_output = `#{cmd}`

  if $?.exitstatus != 0
    STDERR.puts "Error: nmap command failed"
    STDERR.puts "Make sure nmap is installed: brew install nmap"
    exit 1
  end

  parse_nmap_xml(xml_output)
end

def parse_nmap_xml(xml_string)
  doc = REXML::Document.new(xml_string)
  devices = []

  # Iterate through each <host> element
  doc.elements.each('nmaprun/host') do |host|
    status = host.elements['status']
    next unless status && status.attributes['state'] == 'up'

    device = {}

    # Extract IPv4 address
    addr = host.elements['address[@addrtype="ipv4"]']
    device[:ip] = addr.attributes['addr'] if addr

    # Extract MAC address
    mac = host.elements['address[@addrtype="mac"]']
    if mac
      device[:mac] = mac.attributes['addr']
      device[:vendor] = mac.attributes['vendor'] || 'Unknown'
    end

    # Extract hostname (if available)
    hostname = host.elements['hostnames/hostname']
    device[:hostname] = hostname.attributes['name'] if hostname

    devices << device
  end

  devices
end

if ARGV.length != 1
  STDERR.puts "Usage: #{$0} <CIDR>"
  STDERR.puts "Example: #{$0} 192.168.1.0/24"
  exit 1
end

cidr = ARGV[0]

# Basic CIDR format validation
unless cidr =~ /^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\/\d{1,2}$/
  STDERR.puts "Error: Invalid CIDR format"
  STDERR.puts "Expected format: 192.168.1.0/24"
  exit 1
end

devices = scan_network(cidr)

puts JSON.generate(devices)
