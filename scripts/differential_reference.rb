# frozen_string_literal: true

# Renders the differential corpus with Shopify Liquid and writes the results
# that integration/differential_test.go compares liquidgo with.
#
# Usage, from the root of the repository, with a checkout of Shopify Liquid
# v5.11.0 (its gems strscan >= 3.1.1 and bigdecimal installed):
#
#   TZ=UTC ruby -I <shopify-liquid>/lib scripts/differential_reference.rb         # write
#   TZ=UTC ruby -I <shopify-liquid>/lib scripts/differential_reference.rb --check # compare
#
# The corpus is integration/testdata/differential/templates.json, written by
# hand, plus generated templates: fragments of tags, filters and delimiters
# joined at random from a fixed seed, so that every run writes the same file.

require "json"
require "liquid"

DIR = File.expand_path("../integration/testdata/differential", __dir__)
EXPECTED = File.join(DIR, "expected.json")
MODES = %w[lax strict].freeze
GENERATED = 400

FRAGMENTS = [
  "{%", "%}", "{{", "}}", "{%-", "-%}", "{{-", "-}}", " if x ", " endif ", " for i in arr ", " endfor ",
  " raw ", " endraw ", " comment ", " endcomment ", "{% liquid\n", "\n", " # ", " else ", " unless x ",
  " endunless ", " capture c ", " endcapture ", " assign x = ", " | ", " upcase ", ": ", ",", "[", "]",
  "'", "\"", "x", "arr", "1", " and ", " == ", "a b", "<p>", "%", "}", "{", "c", " echo ", "{%raw%}",
  "{%endraw%}", "{% if true %}", "{% endif %}", " size ", " first ", " (1..3) ", " cycle 'a', 'b' ",
  " tablerow i in arr ", " endtablerow ", " case x ", " when 'a' ", " endcase ", " -1 ", " 2.5 ",
  " nothing ", " h.k ", " s ", " slice: 1 ", " truncate: 3 ", " append: ", " split: ',' ", " join: '-' ",
  " default: ", " plus: 1 ", " divided_by: 0 ", " escape ", " increment c ", " break ", " continue ",
].freeze

def generated
  rng = Random.new(11)
  (1..GENERATED).map do |i|
    template = Array.new(rng.rand(1..10)) { FRAGMENTS[rng.rand(FRAGMENTS.size)] }.join
    { "id" => format("generated-%03d", i), "template" => template }
  end
end

def render(template, data, mode)
  parsed = Liquid::Template.parse(template, error_mode: mode.to_sym)
  output = parsed.render(Marshal.load(Marshal.dump(data)))
  { "output" => output.dup.force_encoding("UTF-8").scrub("�") }
rescue Liquid::Error => e
  { "error" => "#{e.class}: #{e.message}" }
rescue StandardError, SystemStackError => e
  { "error" => "ruby #{e.class}" }
end

templates = JSON.parse(File.read(File.join(DIR, "templates.json"), encoding: "UTF-8")) + generated
data = JSON.parse(File.read(File.join(DIR, "data.json"), encoding: "UTF-8"))
results = templates.flat_map do |t|
  MODES.map { |mode| { "id" => t["id"], "mode" => mode, "template" => t["template"] }.merge(render(t["template"], data, mode)) }
end
text = JSON.pretty_generate({ "liquid" => Liquid::VERSION, "results" => results }) + "\n"

if ARGV.include?("--check")
  if File.binread(EXPECTED) != text.b
    warn "#{EXPECTED} is not what Shopify Liquid #{Liquid::VERSION} renders: run this script without --check"
    exit 1
  end
else
  File.binwrite(EXPECTED, text)
end
