# Renders hostile template N with Shopify Liquid and its resource limits, as
# `libcmp hostile` does; the caller bounds the time and the memory.
# ruby -I LIQUID/lib hostile.rb templates.json N
require "json"
require "liquid"

t = JSON.parse(File.read(ARGV[0]))
Liquid::Environment.default.default_resource_limits = t["limits"].transform_keys(&:to_sym)
begin
  template = Liquid::Template.parse(t["hostile"][Integer(ARGV[1])])
rescue Liquid::Error => e
  puts "refused: #{e.message[0, 80]}"
  exit
end
begin
  out = template.render!({})
  if template.resource_limits.reached?
    puts "stopped: #{out[0, 80]}"
  else
    puts "rendered: #{out.bytesize} bytes"
  end
rescue Liquid::Error => e
  puts "stopped: #{e.message[0, 80]}"
end
